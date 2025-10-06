FROM docker.io/mongo:latest
RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y mongodb-database-tools
# test if the tools are installed
RUN mongodump --version && mongorestore --version
