# ECR Cleaner
This tool helps you clean up your ECR repository.  ECR is limited in how you can remove images by either a amount of elapsed time or via an expression matching the name.  Due to the limitations in this combination this tool exists.

### Purging
* The image must be over 90 days old unless `--days` is set.
* Image tags that are [SemVer](https://semver.org/) compatible are kept by default set `--no-semver` to remove SemVer tags.  If the tag starts with a r or v those characters will be stripped and the remainder of the name will be tested for SemVer compatiability.
* `--dry-run` allows you to run the command to see what the results would be.
* By default it will skip any repo that would have all images removed set `--force` to remove them anyways

## Helm Install

Using helm 3

```
$  helm repo add sstarcher https://shanestarcher.com/helm-charts/
$ helm install ecr-cleaner sstarcher/ecr-cleaner
```

If using kube2iam you will want to specify podAnnotations
```yaml
podAnnotations:
    iam.amazonaws.com/role: IAM_ROLE
```

## Automated Docker Builds

This is handled by docker hub's automated [build process](https://hub.docker.com/repository/docker/sstarcher/ecr-cleaner/).
