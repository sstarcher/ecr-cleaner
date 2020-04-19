package cleaner

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	log "github.com/sirupsen/logrus"
)

// Cleaner allows for pruning ECR images
type Cleaner interface {
	Prune(time.Duration, bool, bool, bool, string) error
}

type cleaner struct {
	client *ecr.ECR
}

// New Cleaner for pruning ECR
func New(region *string) (Cleaner, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if region != nil && *region != "" {
		sess.Config.WithRegion(*region)
	} else if sess.Config.Region == nil || *sess.Config.Region == "" {
		sess.Config.WithRegion("us-east-1")
		region := ""
		svc := ec2metadata.New(sess)
		if svc.Available() {
			log.Debug("loading region from metadata service")
			var err error
			region, err = svc.Region()
			if err != nil {
				log.Warnf("unable to fetch region %v", err)
			} else {
				log.Infof("using region from metadata service %s", region)
			}
		}
		sess.Config.WithRegion(region)
	}

	if sess.Config.Region == nil || *sess.Config.Region == "" {
		return nil, errors.New("region must be specified")
	}
	return &cleaner{
		client: ecr.New(sess),
	}, nil
}

// Prune the ECR registry
func (c *cleaner) Prune(age time.Duration, semanticVersioning bool, dryRun bool, force bool, repo string) error {
	var repos []*ecr.Repository
	if repo == "" {
		var err error
		repos, err = c.repos()
		if err != nil {
			return err
		}
	} else {
		repos = append(repos, &ecr.Repository{RepositoryName: &repo})
	}

	var totalSize int64
	for _, repository := range repos {
		repo := *repository.RepositoryName
		log := log.WithField("repo", repo)
		imgs, err := c.images(repo)
		if err != nil {
			log.Warnf("Failed to fetch %v", err)
			break
		}

		if len(imgs) == 0 {
			log.Info("no images")
			continue
		}

		var prune []*ecr.ImageDetail
		for _, image := range imgs {
			if image != nil {
				elapsed := time.Now().Sub(*image.ImagePushedAt)
				if elapsed > age {
					protected := false
					for _, tag := range image.ImageTags {
						if *tag == "latest" {
							protected = true
							break
						}

						if semanticVersioning {
							version := strings.TrimLeft(*tag, "rv")
							_, err := semver.NewVersion(version)
							if err == nil {
								log.WithField("tag", *tag).Debugf("protected")
								protected = true
								break
							}
						}
					}

					if !protected {
						prune = append(prune, image)
					}
				}

			} else {
				prune = append(prune, image)
			}
		}

		var size int64
		ids := make([]*ecr.ImageIdentifier, len(prune))
		for index, item := range prune {
			size += *item.ImageSizeInBytes / 1024 / 1024
			ids[index] = &ecr.ImageIdentifier{
				ImageDigest: item.ImageDigest,
			}
		}

		totalSize += size
		total := len(imgs)
		pruneCount := len(prune)
		remainder := total - pruneCount
		log.WithField("total", total).
			WithField("prune", pruneCount).
			WithField("remainder", remainder).
			WithField("size", fmt.Sprintf("%dmb", size)).
			Infof("")
		if remainder == 0 && !force {
			log.Warnf("skipping as no images would remain, set --force")
		} else if !dryRun {
			batch := 100
			for i := 0; i < len(ids); i += batch {
				j := i + batch
				if j > len(ids) {
					j = len(ids)
				}

				_, err := c.client.BatchDeleteImage(&ecr.BatchDeleteImageInput{
					ImageIds:       ids[i:j],
					RepositoryName: &repo,
				})
				if err != nil {
					log.WithError(err).Error("failed to batch delete images")
				}
			}
		}
	}
	log.Infof("total size %dgb", totalSize/1024)
	return nil
}

func (c *cleaner) repos() ([]*ecr.Repository, error) {
	var repos []*ecr.Repository
	input := &ecr.DescribeRepositoriesInput{}
	for {
		result, err := c.client.DescribeRepositories(input)
		if err != nil {
			return nil, err
		}

		repos = append(repos, result.Repositories...)
		if result.NextToken == nil {
			break
		}
		input.SetNextToken(*result.NextToken)
	}

	return repos, nil
}

func (c *cleaner) images(repo string) ([]*ecr.ImageDetail, error) {
	input := &ecr.ListImagesInput{
		RepositoryName: aws.String(repo),
	}

	images := []*ecr.ImageDetail{}
	for {
		result, err := c.client.ListImages(input)
		if err != nil {
			return nil, err
		}

		if len(result.ImageIds) > 0 {
			desc, err := c.client.DescribeImages(&ecr.DescribeImagesInput{
				RepositoryName: &repo,
				ImageIds:       result.ImageIds,
			})
			if err != nil {
				return nil, err
			}

			images = append(images, desc.ImageDetails...)
		}

		if result.NextToken == nil {
			break
		}
		input.SetNextToken(*result.NextToken)
	}

	return images, nil
}
