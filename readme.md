# Anabel Hospitality IoT Alert System REST API Part 1

## Prerequisites
1. [Golang](https://golang.org/dl/)
2. [MongoDB](https://docs.mongodb.com/manual/administration/install-community/)
3. [AWS S3](https://aws.amazon.com/s3/getting-started/)
4. [AWS SES](https://aws.amazon.com/ses/)

# AWS S3 Setup
Follow the [document](https://docs.aws.amazon.com/AmazonS3/latest/user-guide/create-configure-bucket.html) and purchase and setup s3 bucket

# AWS SES Setup
Follow the [document](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/send-email-set-up.html) and purchase and setup Simple Email Service

## Configuration

Following items are configurable from `config.sample.yaml`

| key                                     | description                                       |
| --------------------------------------- | --------------------------------------------------|
| server.listen                           | the server address to run go web application      |
| mongodb.url                             | the mongo db url to connect                       |
| aws.access_key_id                       | the aws public key                                |
| aws.secret_access_key                   | the aws secret key                                |
| aws.s3_region                           | the aws s3 region                                 |
| aws.s3_bucket                           | the aws s3 bucket name                            |
| app.token_validation_period_in_minutes  | application token validation period               |
| app.forntend_url                        | application front end app url                     |
| email.sender                            | the email sender address                          |
| email.activation_subject                | the email activation subject                      |
| email.activation_body                   | the email activation body                         |
| log.file                                | the log file                                      |
| log.level                               | the log level                                     |


# Run locally
- Copy `config.sample.yaml` to `config.yaml` and update it accordingly
  - For development no need to change mongodb host and port. DB name can be changed
- To build the container `docker-compose -f .\docker-compose.dev.yml build`
- For Running `docker-compose -f .\docker-compose.dev.yml up`
- For Close <kbd>Ctrl</kbd> + c and `docker-compose down`


# Run in Prod Mode
- Start your mongo db or Get the Mongodb info from provider
- Copy `config.sample.yaml` to `config.yaml` and update it accordingly
- Change the port in `docker-compose.yml` accordingly.
- Run `docker-compose up --build` to build and run.
- Need to open specific port on the host system. Like
  - `ufw allow 4201/tcp`
  - `ufw allow 4001/tcp`
  - `ufw enable`

## Postman scripts

- postman scripts are inside `/docs/postman`
- All end poins are listed on postman file with example

## Other informations

1. Updated the swagger openapi file for few api located ar `/docs/swagger`
2. 1 Api exposed under the package dummy to insert Admin user