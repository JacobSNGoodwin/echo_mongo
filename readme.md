# Echo-Mongo-GCStorage Demo

This project was created to demonstrate building a basic REST API with the [Echo Framework](https://echo.labstack.com/guide). Mongo DB is used as a database, and Google Cloud Storage is used for storing file uploads from multipart/form data.

# Usage

* Inside of the directory where you store this repository, start the database by running
  > docker-compose up
* Run the main.go file with command-line arguments. To do this, I have created an example bash script file, [run_ex.sh](run_ex.sh). You can also create a command or powershell script on Windows in a like-manner. Note that this command builds an executable file. This helps when developing on Windows to avoid pesky firewall warnings each time the server is restarted. 
* The example script shows 3 command line arguments/flags that you need to set.
  1. -dburi : this is the mongo db uri, and defaults to mongodb://root:example@localhost:27017
    * This parameter does not need to be set if you use the default MongoDB settings in the docker-compose file.
  2. -gcconfig : this holds the path the the json file with your Google cloud config
    * Instructions for getting this JSON file and setting up a file storage bucket can be found in [Google Cloud's documentation](https://cloud.google.com/storage/docs/reference/libraries#client-libraries-install-go)
  3. -gcbucket : this holds the name of your file storage bucket on Google Cloud
* Test routes with client or program of your choice (ie, [Postman](https://www.getpostman.com/))
  * Available routes are listed in [routes.json](routes.json)

# Todo
* Document endpoints - [Swagger](https://github.com/swaggo/swag)?
* Implement validation on request bodies
* Check response types and error handling
* Create a GO client to test endpoints [as shown here](https://echo.labstack.com/guide/testing)