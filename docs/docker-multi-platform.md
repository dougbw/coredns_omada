## Building multi-platform docker images
In order to build multi platform docker images there are some additional setup is required

* In your docker engine configuration file set:
``` 
"experimental": true,
  "features": {
    "buildkit": true
}
```
* run the following commands
```
docker buildx create --name multiplatform
docker buildx use multiplatform
docker buildx inspect --bootstrap
```

# References
* https://www.docker.com/blog/multi-arch-images/
* https://cloudolife.com/2022/03/05/Infrastructure-as-Code-IaC/Container/Docker/Docker-buildx-support-multiple-architectures-images/
