# Uploading Bootstrap Tarball

1. Download a CentOS-Stream-Container-Base tarball with a timestamp in its name, like `20230704`, from:
   - [https://cloud.centos.org/centos/9-stream/x86_64/images/](https://cloud.centos.org/centos/9-stream/x86_64/images/)
   - [https://cloud.centos.org/centos/9-stream/aarch64/images/](https://cloud.centos.org/centos/9-stream/aarch64/images/)
2. Upload them to artifactory in the subpath `eext-sources/bootstrap/CentOS-Stream/`
```
  curl -H "Authorization: Bearer ${ARTIFACTORY_TOKEN}" -X PUT https://artifactory.infra.corp.arista.io/artifactory/eext-sources/bootstrap/CentOS-Stream/ -T <TARBALL_PATH>
```
3. Update the `CHECKSUM` file in the local repo for the new entries from the `CHECKSUM` files:
   - [https://cloud.centos.org/centos/9-stream/x86_64/images/CHECKSUM](https://cloud.centos.org/centos/9-stream/x86_64/images/CHECKSUM)
   - [https://cloud.centos.org/centos/9-stream/aarch64/images/CHECKSUM](https://cloud.centos.org/centos/9-stream/aarch64/images/CHECKSUM)
4. Update the `EEXT_BOOTSTRAP_VERSION` environment variable in `barney.yaml`.
