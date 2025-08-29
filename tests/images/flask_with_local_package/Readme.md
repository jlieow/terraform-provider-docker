Follows: https://cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-python-service

# Navigate to root folder where Dockerfile is located and build a docker image

TAG=sample-nested
docker build -t $TAG . --platform linux/amd64
docker run -e PORT=8080 -it -p 8000:8080 --name nested -d "${TAG}:latest"

Connect to `http://localhost:8000/api/hello-api`

# Configures Docker to authenticate to Artifact Registry

gcloud auth configure-docker asia-southeast1-docker.pkg.dev

# Build Image

docker build -t test/build:latest .
docker run -e PORT=8080 -it -p 8000:8080 --name container -d test/build:latest
docker exec -it container sh

# Image Tag

Add tag
`docker image tag just/a:test just/a:try`

Remove tag
`docker rmi just/a:try`
`docker rmi just/for:fun`

# Terraform Log

TF_LOG=TRACE TF_LOG_PATH=trace.txt terraform plan
