# Generated files

Files in `types/models` were generated and slightly modified to:

* Minimize unused fields
* Change package names

docker run -v $PWD/pkg/generated:/output/ swaggerapi/swagger-codegen-cli generate -l go -o /output -i https://raw.githubusercontent.com/square/connect-api-specification/master/api.json

Files in `types/webhooks` were manually created.
