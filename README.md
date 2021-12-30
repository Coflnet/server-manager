# server-manager

service to create and destroy google cloud servers

## api

### list

GET /list

list all servers

UserId is the same as given in kafka message or api request

authentication_token are 20 random characters

status can be `creating`, `ok`, `deleting`

example response
```
[
	{
		"ID": "61cdad3bb9c351e933ff6c58",
		"type": "c2-standard-4",
		"name": "dpdreurwjp",
		"ip": "",
		"status": "creating",
		"UserId": "userIdExample",
		"authenticationToken": "txieaweukxavtpzrgdwf",
		"InstanceId": {},
		"createdAt": "2021-12-30T12:59:39.139Z",
		"plannedShutdown": "2021-12-30T12:53:43.469Z"
	}
]
```

### create

POST /create

create a new server, planned_shutdown and userId are optional but should be set

example request
```
{
	"type": "e2-micro",
	"plannedShutdown": "2021-12-24T19:14:43.469947584Z".
  	"userId": "random-string"
}
```

### destroy

DELETE /destroy

destroy a server by id

example request
```
{
	"id": "61c5bb3a5556b53b72371a20"
}
```

### update

PATCH /update

`id` of the server that should be updated
`planned_shutdown` to modify the planned shutdown

example request
```
{
	"id": "619031ce1640d001136a4345",
	"plannedShutdown": "2022-11-13T19:14:43.469947584Z"
}
```

### metrics

prometheus metrics available on port 3001 /metrics

## additional information

three env vars are set for the running container in gcp

first is `SNIPER_DATA_USERNAME` the second env var is `SNIPER_DATA_PASSWORD`

third is `MOD_AUTHENTICATION_TOKEN` random token to authenticate mod connections

this variables should be used to download sniper data from the "main" cluster.
