package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tiiuae/rclgo/pkg/rclgo"
	"github.com/tiiuae/ros2msg"
	"github.com/tiiuae/ros2msg/transport"
)

func rootCmd(rclArgs *rclgo.Args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ros2msg",
		Short: "Command line utility for transporting ROS 2 messages",
	}
	verbose := cmd.PersistentFlags().BoolP("verbose", "v", false, "Control verbosity of the command")
	cmd.AddCommand(publishCmd(verbose, rclArgs))
	return cmd
}

func publishCmd(verbose *bool, rclArgs *rclgo.Args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish messages from a file to the local ROS network",
		Long: `Publish messages from a file to the local ROS network.

In addition to the options specified below, standard ROS arguments are also supported.`,
	}
	src := ""
	cmd.Flags().StringVarP(&src, "source", "s", src, "Source file where the message stream is read from. If empty, defaults to standard input.")
	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		publishers := make(map[string]*rclgo.Publisher)
		rclctx, err := rclgo.NewContext(0, rclArgs)
		if err != nil {
			return fmt.Errorf("failed to create ROS context: %v", err)
		}
		defer rclctx.Close()
		node, err := rclctx.NewNode("ros2msgtransport", "")
		if err != nil {
			return fmt.Errorf("failed to create ROS node: %v", err)
		}

		inFile := os.Stdin
		if src != "" {
			inFile, err = os.Open(src)
			if err != nil {
				return fmt.Errorf("failed to open source: %v", err)
			}
			defer inFile.Close()
		}
		inReader := bufio.NewReader(newCancelableReader(cmd.Context(), inFile))

		var msg ros2msg.Message
		for {
			err = transport.ReadProtoStream(inReader, &msg)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
					return nil
				}
				fmt.Fprintln(os.Stderr, "failed to read message:", err)
				continue
			}
			if *verbose {
				fmt.Fprintln(os.Stderr, "received message:", &msg)
			}
			pubKey := msg.Topic + "\u0000" + msg.Type
			pub := publishers[pubKey]
			if pub == nil {
				parts := strings.Split(msg.Type, "/")
				if len(parts) != 3 {
					fmt.Fprintln(os.Stderr, "invalid topic type:", msg.Type)
					continue
				} else if parts[1] != "msg" {
					fmt.Fprintf(os.Stderr, "non-msg type %s is not supported\n", msg.Type)
					continue
				}
				ts, err := rclgo.LoadDynamicMessageTypeSupport(parts[0], parts[2])
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to load type support for %s: %v\n", msg.Type, err)
					continue
				}
				pub, err = node.NewPublisher(msg.Topic, ts, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, "failed to create publisher:", err)
					continue
				}
				publishers[pubKey] = pub
			}
			if err = pub.PublishSerialized(msg.Data); err != nil {
				fmt.Fprintln(os.Stderr, "failed to publish message:", err)
			}
		}
	}
	return cmd
}

type readResult struct {
	err error
	n   int
}

type cancelableReader struct {
	//nolint:containedctx
	context    context.Context
	src        io.Reader
	readChan   chan []byte
	resultChan chan readResult
}

func newCancelableReader(ctx context.Context, src io.Reader) *cancelableReader {
	r := &cancelableReader{
		context:    ctx,
		src:        src,
		readChan:   make(chan []byte),
		resultChan: make(chan readResult),
	}
	go r.reader()
	return r
}

func (r *cancelableReader) reader() {
	var result readResult
	for {
		select {
		case <-r.context.Done():
			return
		case buf := <-r.readChan:
			result.n, result.err = safeRead(r.src, buf)
		}
		select {
		case <-r.context.Done():
			return
		case r.resultChan <- result:
		}
	}
}

func (r *cancelableReader) Read(buf []byte) (int, error) {
	select {
	case <-r.context.Done():
		return 0, r.context.Err()
	default:
	}
	select {
	case <-r.context.Done():
		return 0, r.context.Err()
	case r.readChan <- buf:
	}
	select {
	case <-r.context.Done():
		return 0, r.context.Err()
	case result := <-r.resultChan:
		return result.n, result.err
	}
}

type panicError struct{ Value interface{} }

func (e panicError) Error() string { return fmt.Sprint(e.Value) }

func safeRead(r io.Reader, buf []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = panicError{Value: r}
		}
	}()
	return r.Read(buf)
}

func run() error {
	rclArgs, restArgs, err := rclgo.ParseArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse ROS arguments: %v", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	cmd := rootCmd(rclArgs)
	cmd.SetArgs(restArgs)
	return cmd.ExecuteContext(ctx)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
