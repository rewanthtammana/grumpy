# grumpy with cosign validation

This code allows to use cosign with validating admission controllers for verifying the integrity of images.

## Build from scratch

1. `cosign generate-key-pair && cat cosign.pub | base64 -w0` - Copy this text to kubernetes secret, `public` in `_manifest_.yaml`
1. Build the docker image from scratch `docker build . -t $IMAGENAME && docker push $IMAGENAME` or use `rewanthtammana/build:cosign`
2. Generate certificates & perform deployments with, `./deploy.sh`
3. Check the status
4. I have already signed an image & pushed it to my dockerhub. For validation run,
    1. Deploy Singed Image - `kubectl run --rm -it rewanthtammana/python:alpine`
    2. Deploy Unsigned Image - `kubectl run --rm -it nginx`
