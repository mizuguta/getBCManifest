# getBCManifest

Demo/Utility web app for Video Cloud Delivery Rules. 
It may be useful to get rendition list for assets by content-aware encodings.

## How to run
1. Create a new envfile under envfiles/ directory. It must have a Brightcove Video Cloud account information. File surfix must be *.env
2. Update GO_ENV value in docker-compose.yml to the name of the envfile you created in step 1.
  You don't need to include .env suffix
3. % docker-compose build
4. % docker-compose up
5. You can access http://localhost:8321

## How to use
http://localhost:8321/getManifest/{video_id}/{config_name}
- config_name is optional
- you can see all available config_names when you access to http://localhost:8321/getManifest/

