package main

const IncorrectUserInputMessage string = `Your request was denied.
The format of your request is incorrect.

For help see: cf pcc --help
`

const NoServiceKeyMessage string = `Please create a service key for %s.
To create a key enter: 

	cf create-service-key %s <your_key_name>
	
For help see: cf create-service-key --help
`
const GenericErrorMessage string = `Cannot retrieve credentials. Error: %s`
const InvalidServiceKeyResponse string = `The cf service-key response is invalid.`
const ProvidedUsernameAndNotPassword string = `You did not specify your password.
Please enter username and password:

	cf pcc %s %s -u=%s -p=<your_password>

For help see: cf pcc --help
`
const ProvidedPasswordAndNotUsername string = `You did not specify your username.
Please enter username and password:

	cf pcc %s %s -u=<your_username> -p=%s

For help see: cf pcc --help
`
const NeedToProvideUsernamePassWordMessage string = `You need to provide your username and password.
The proper format is: cf pcc %s %s -u=<your_username> -p=<your_password>

For help see: cf pcc --help
`

const NoRegionGivenMessage string = `You need to provide a region (-r flag) to execute your command.

To see your available regions:

	cf pcc %s list regions

For help see: cf pcc --help
`

const NoIDGivenMessage string = `An identifier is required for all get commands.

Please re-enter your command appended with -id=<your_object_of_interest>
`
const NoJsonFileProvidedMessage string = `A JSON configuration file is required for all create/post commands.

Please re-enter your command appended with -d=<your_json_configuration_file>`

const NoEndpointFoundMessage string = `No endpoint was found for your request.

For help see: cf pcc --help
`


const Ellipsis string = "…"
