/*
 *
 * Copyright 2015, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *server) HighFive(stream pb.Greeter_HighFiveServer) error {
	fmt.Println("HighFive server begin ...")
	r, w, _ := os.Pipe()
	read, write, _ := os.Pipe()

	go func() {
		// Recv loop
		for {
			req, err := stream.Recv()
			fmt.Printf("stream received: %s\n", req.Content)
			if err == io.EOF {
				return
			}
			if err != nil {
				panic(err)
			}
			w.Write(req.Content)
			fmt.Println("content wrote")
			defer w.Close()
		}
	}()

	go func() {
		myService(r, write)
	}()

	// Send loop
	for {
		reader := bufio.NewReader(read)
		data, _, _ := reader.ReadLine()
		fmt.Println("content read")
		reply := &pb.HighReply{Content: data}
		if err := stream.Send(reply); err != nil {
			panic(err)
		}
		fmt.Println("content sent")
	}

	return nil
}

// The real service, it's a echo loop
func myService(stdin io.ReadCloser, stdout io.WriteCloser) error {
	for {
		reader := bufio.NewReader(stdin)
		line, _, err := reader.ReadLine()
		fmt.Printf("line readed: %s\n", line)
		if err != nil {
			return err
		}
		if string(line[:]) == "exit" {
			return nil
		} else {
			// Need to add a "\n" at the end of string, otherwise it will deadlock!
			strMsg := fmt.Sprintf("Processed: %s\n", line)
			stdout.Write([]byte(strMsg))
			fmt.Println("line wrote")
		}
	}
	/**
	buf := make([]byte, 32*1024)
	for {
		nr, er := stdin.Read(buf)
		if nr > 0 {
			strMsg := fmt.Sprintf("Processed: %s\n", buf[0:nr])
			_, ew := stdout.Write([]byte(strMsg))
			if ew != nil {
				return ew
			}
		}
		if er == io.EOF {
			return er
		}
		if er != nil {
			return er
		}
	}
	**/
	return nil
}

func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {

	return written, err
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	s.Serve(lis)
}
