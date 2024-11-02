Follows: https://cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-python-service


# Navigate to root folder where Dockerfile is located and build a docker image
docker build -t react-app . --platform linux/amd64
docker tag react-app asia-southeast1-docker.pkg.dev/sg-rd-ce-jerome-lieow/cloud-run-source-deploy/react-app

# Configures Docker to authenticate to Artifact Registry
gcloud auth configure-docker asia-southeast1-docker.pkg.dev

# Push image to 
docker push asia-southeast1-docker.pkg.dev/sg-rd-ce-jerome-lieow/cloud-run-source-deploy/react-app

# Docker run
docker run -d --restart=always -p 3000:3000 asia-southeast1-docker.pkg.dev/sg-rd-ce-jerome-lieow/cloud-run-source-deploy/react-app:latest

docker run -d --restart=always -p 3000:3000 react-app:latest