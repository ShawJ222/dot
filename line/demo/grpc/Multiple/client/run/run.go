/*
@Time :  2019/3/7 17:43
@Author : niko
@File : run
@Software: GoLand
*/
package main

import (
	"fmt"
	"github.com/scryinfo/dot/dot"
	"github.com/scryinfo/dot/dots/grpc/client"
	pb "github.com/scryinfo/dot/line/demo/pb"
	"github.com/scryinfo/dot/line/lineimp"
)

var f gclient.GrpcClienter
var t gclient.GrpcClienter

func init() {
	l := lineimp.New()
	l.ToLifer().Create(nil)

	gclient.Add(l, dot.LiveId("dd05cbec-e3d0-4be3-a7df-87b0522ac46a"))
	gclient.Add(l, dot.LiveId("dd05cbec-e3d0-4be3-a7df-87b0522ac46b"))

	err := l.Rely()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = l.CreateDots()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = l.ToLifer().Start(false)
	if err != nil {
		fmt.Println(err)
		return
	}

	{
		d, _ := l.ToInjecter().GetByLiveId("dd05cbec-e3d0-4be3-a7df-87b0522ac46a")
		f = d.(gclient.GrpcClienter)
	}

	{
		d, _ := l.ToInjecter().GetByLiveId("dd05cbec-e3d0-4be3-a7df-87b0522ac46b")
		t = d.(gclient.GrpcClienter)
	}
}

func A() {
	conn := f.GetConn()

	ctx := f.GetCtx()

	//用户实现 start
	c := pb.NewTestClient(conn)

	c1, err := c.SayHello(ctx, &pb.TestRequest{Name: "shrimpliaoA"})

	fmt.Println("err", err)

	fmt.Printf("@@@c1: %s", c1.Message)

	//用户实现 end

	f.Stop(false)
	f.Destroy(false)
}

func B() {
	conn := t.GetConn()

	ctx := t.GetCtx()

	//用户实现 start
	c := pb.NewTestClient(conn)

	c1, err := c.SayHello(ctx, &pb.TestRequest{Name: "shrimpliaoB"})

	fmt.Println("err", err)

	fmt.Printf("@@@c1: %s", c1.Message)

	//用户实现 end

	t.Stop(false)
	t.Destroy(false)
}

func main() {
	A()
	B()
}