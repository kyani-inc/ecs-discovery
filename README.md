# ecs-discovery

ecs-discovery pulls standard metadata about containers running on ECS using the ECS API.

#### Metadata used

The following is information used to create the metadata for discovering services:

- Name : Task Name as in ECS
- TaskARN
- IPAddress
- ContainerPort
- HostPort
- DockerLabel.hostname : In the ECS Task definition you can add DockerLabels, if you have a DockerLabel named `hostname` it will override the default domain and expose that instead of the default `name.domain`

#### Notes

Although rough around the edges, the ConsulKV implementation is being used internally by us to manage/discover several hundred production containers across multiple servers.

## Todo

- Tests
- Better Documentation

## Contributing

Contributions are welcome! Please submit a pull request and we'll discuss any changes and approve the request.

Providers are encouraged such as etcd. We will work (where we can) with you on assistance.

## License

MIT 