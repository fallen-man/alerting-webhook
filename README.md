having `bypass-filter` in `keys` list of `config.yml` will ignore all the list and simply just turn all the incoming request body to text

* target webhook url is hardcoded and needs to be changed in code
* request Json Structure is Hardcoded and needs to be change in code if you want to use this code with another target webhook
* source Json have no Structure limitations but needs to be single layer (nested Json not supported)

* `limitations will remove in future versions`