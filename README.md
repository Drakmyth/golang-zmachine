# golang-zmachine

This is a Z-Machine interpreter implemented in Go. It was built as a learning exercise to both gain a better handle on Go itself, as well as to gain insight into the structure and operation of a text adventure game engine. Feature-wise it is pretty barebones comared to other Z-Machine implementations, but it gets the job done.

### Built With

* [![Golang][golang-shield]][golang-url]

## Usage

```sh
> zmachine <story-path>
```

Arguments      | Description
-------------- | -----------
`<story-path>` | Load and play the specified story file

Execute `zmachine help` for more detailed information.

## Development

### Build

```sh
> go build -o zmachine
```

or on Windows

```sh
> go build -o zmachine.exe
```

### Debugging

Using the [Delve][delve-url] debugger with CLI applications is a little tricky. See the [Delve documentation][delve-debug-url] for recommended procedures on how to do this. A VSCode [launch.json](./.vscode/launch.json) has been provided that runs and debugs the build using a hardcoded story file path.

### Release

While `zmachine` is a CLI application and is intended to be executed by directly, a containerized installation is also provided. This container utilizes a dedicated build stage along with the [scratch][scratch-url] Docker image to ensure the final image contains only the necessary resources and nothing else.

#### Build

```sh
> docker build . -t zmachine
> docker run zmachine <story-path>
```

## License

Distributed under the MIT License. See [LICENSE.md](./LICENSE.md) for more information.


<!-- Reference Links -->
[golang-url]: https://go.dev
[golang-shield]: https://img.shields.io/badge/golang-09657c?style=for-the-badge&logo=go&logoColor=79d2fa
[delve-url]: https://github.com/go-delve/delve
[delve-debug-url]: https://github.com/go-delve/delve/blob/master/Documentation/faq.md#-how-can-i-use-delve-to-debug-a-cli-application
[scratch-url]: https://hub.docker.com/_/scratch/