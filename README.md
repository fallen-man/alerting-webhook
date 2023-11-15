having `bypass-filter` in `keys` list of `config.yml` will ignore all the list and simply just turn all the incoming request body to text
* example of docker run command:
    * `dcoker run -d --name alerting-webhook -p 7777:7777 -v ./config.yml:/go/bin/config.yml <image name>`
* target webhook url is hardcoded and needs to be changed in code
* request Json Structure is Hardcoded and needs to be change in code if you want to use this code with another target webhook
* source Json have no Structure limitations but needs to be single layer (nested Json not supported)

* docker image on local repo:
    * `docker.mofid.dev/mofidonline/alerting-webhook/alerting-webhook:0.0.1`
    * please consider using same address and name with a diffrent version if you want to push a custom built image from this project to repository

* `limitations will remove in future versions`