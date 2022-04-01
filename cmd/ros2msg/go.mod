module github.com/tiiuae/ros2msg/cmd/ros2msg

go 1.18

require (
	github.com/spf13/cobra v1.4.0
	github.com/tiiuae/rclgo v0.0.0-20220318150403-b3b61182b252
	github.com/tiiuae/ros2msg v0.0.0-00010101000000-000000000000
)

replace github.com/tiiuae/ros2msg => ../..

require (
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
