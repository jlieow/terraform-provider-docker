# Use an official Node runtime as a parent image
FROM node:20.13.1-alpine

# Set the working directory to /app
ENV APP_HOME /app
WORKDIR $APP_HOME

# Copy the files to the working directory
COPY . ./

# Install the dependencies
RUN npm install

# Start the application
CMD npm start