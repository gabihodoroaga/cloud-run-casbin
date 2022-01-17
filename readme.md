# cloud-run-casbin

The is an example application of how you can use "Sign-in with Google" and Casbin 
to simplify identity and add authentication and authorization to a web application.
You can read more in this blog post [Simplified identity with 'Sign-in with Google' and Casbin](http://hodo.dev/posts/post-33-gcp-simplified-identity/)

## How to run locally 

### Setup database

Clone the repo:

```bash
git clone https://github.com/gabihodoroaga/cloud-run-casbin.git
cd cloud-run-casbin
```

Create the PostgreSQL server

```bash
docker run --name postgres -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres
```

Create the database

```bash
docker exec postgres psql -U postgres -c "CREATE DATABASE cloudrun_casbin;"

docker exec postgres psql -U postgres cloudrun_casbin \
    -c "CREATE TABLE IF NOT EXISTS users_roles ( \
    id SERIAL PRIMARY KEY, \
    email VARCHAR(100) NOT NULL, \
    role VARCHAR(100) NOT NULL, \
    CONSTRAINT ck_unique_role UNIQUE(email,role));"
```

Add the default admin user
```bash
docker exec postgres psql -U postgres -c "INSERT INTO users_roles (email, role) VALUES ('admin@example.com', 'admin');"
```

Run locally 
```bash
POSTGRESQL_URL='user=postgres password=$POSTGRESQL_PASS database=cloudrun_casbin host=localhost' \
POSTGRESQL_PASS="password" \
GOOGLE_CLIENT_ID="[YOUR-GOOGLE-CLIENT-ID]" \
API_BASE_PATH="http://localhost:8080" \
go run .
```

In order to get a client id follow the instruction provide here: [Get your Google API client ID](https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid)


## How to deploy on Cloud Run

Define variables:

```bash
PROJECT_ID=`gcloud config list --format 'value(core.project)' 2>/dev/null`
INSTANCE_NAME=casbin-test
REGION=us-central1
ZONE=us-central1-a
POSTGRESQL_PASS=SAflsakjw8j
GOOGLE_CLIENT_ID="[YOUR-GOOGLE-CLIENT-ID]"
```

Create SQL instance:

```bash
gcloud sql instances create $INSTANCE_NAME \
    --database-version=POSTGRES_13 \
    --tier=db-g1-small \
    --region=$REGION \
    --root-password=$POSTGRESQL_PASS
```

Run `cloud_sql_proxy` in another terminal (make sure to set the env variables correctly).

```bash
cloud_sql_proxy -instances=$PROJECT_ID:$REGION:$INSTANCE_NAME=tcp:5432
```

Create the database

```bash
PGPASSWORD=$POSTGRESQL_PASS psql -h localhost -p 5432 -U postgres -c "CREATE DATABASE cloudrun_casbin;"
```

Create the tables
```bash
PGPASSWORD=$POSTGRESQL_PASS psql -h localhost -p 5432 -U postgres cloudrun_casbin \
    -c "CREATE TABLE IF NOT EXISTS users_roles ( \
    id SERIAL PRIMARY KEY, \
    email VARCHAR(100) NOT NULL, \
    role VARCHAR(100) NOT NULL, \
    CONSTRAINT ck_unique_role UNIQUE(email,role));"
```

Add the default admin user

```bash
PGPASSWORD=$POSTGRESQL_PASS psql -h localhost -p 5432 -U postgres cloudrun_casbin \
    -c "INSERT INTO users_roles (email, role) VALUES ('admin@example.com', 'admin');"
```

Kill the `cloud_sql_proxy` process in the other terminal window.

Create a secret from the PostgreSQL password

```bash
echo -n "$POSTGRESQL_PASS" | gcloud secrets create postgres-password --data-file=-
```

Add permission for the default service account

```
PROJECT_NUMBER=$(gcloud projects list --filter=$PROJECT_ID --format="value(projectNumber)")
gcloud secrets add-iam-policy-binding postgres-password \
    --member=serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com \
    --role='roles/secretmanager.secretAccessor'

```

Create the docker image for the web app


```bash
gcloud builds submit --tag gcr.io/$PROJECT_ID/webapp --project $PROJECT_ID .
```

Deploy to cloud run

```bash
gcloud run deploy webapp \
--image gcr.io/$PROJECT_ID/webapp \
--add-cloudsql-instances $PROJECT_ID:$REGION:$INSTANCE_NAME \
--set-env-vars=POSTGRESQL_URL="user=postgres password="'$POSTGRESQL_PASS'" database=cloudrun_casbin host=/cloudsql/$PROJECT_ID:$REGION:$INSTANCE_NAME"\
,GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID \
--set-secrets POSTGRESQL_PASS=postgres-password:latest \
--allow-unauthenticated \
--project $PROJECT_ID \
--region $REGION
```

Get the url of the deployed service

```bash
SERVICE_URL=$(gcloud run services describe webapp --platform managed --region $REGION --format 'value(status.url)')
```
Update the API_BASE_PATH environment variable

```bash
gcloud run services update webapp \
    --region $REGION \
    --update-env-vars API_BASE_PATH=$SERVICE_URL
```

Update the "Authorized JavaScript origins" for the client id with the service url

```bash
echo $SERVICE_URL
# currently there is no option to update this using gcloud
# navigate to "APIs & Services"-> "Credentials" -> OAuth 2.0 Client IDs -> Edit 
```

## TODO

[] Add unit tests
