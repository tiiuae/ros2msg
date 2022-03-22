# ros2msg

ros2msg contains Go utilities for processing ROS 2 messages as well as a command
line tool using some of the features.

## ros2msg command line tool

ros2msg CLI tool requires installing and sourcing a ROS galactic environment.
After that the command line tool can be installed by running

    go install github.com/tiiuae/ros2msg/cmd/ros2msg@latest

Make sure that `$GOBIN` is in `$PATH`. Usage information can be displayed by
running `ros2msg --help`.
