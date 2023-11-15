* having `bypass-filter` in `keys` list of `config.yml` will ignore all the list and simply just turn all the incoming request body to text.
* having `bypass-aof` in `keys` list of `config.yml` will ignore saving raw data in a append-only file on disk.
    ** `consider using "bypass-aof" if you don't want to mount a directory for storing AOF files`

* example of docker run command:
    * `dcoker run -d --name alerting-webhook -p 7777:7777 -v ./config.yml:/go/bin/config.yml -v alert-aof:/etc/alerting-webhook <image name>`
* target webhook url is hardcoded and needs to be changed in code
* request Json Structure is Hardcoded and needs to be change in code if you want to use this code with another target webhook
* source Json have no Structure limitations but needs to be single layer (nested Json not supported)

* `limitations will remove in future versions`
