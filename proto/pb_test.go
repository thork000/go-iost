package proto

//go:generate protoc --gofast_out=. *.proto

//func TestRepeat(t *testing.T) {
//	Convey("test of repeat", t, func() {
//		ia := IntArray{I: []int32{1, 2, 3, 4, 5, 6}}
//		buf, err := ia.Marshal()
//		So(err, ShouldBeNil)
//		var ia2 IntArray
//		err = ia2.Unmarshal(buf)
//		So(err, ShouldBeNil)
//		fmt.Println(ia2.I)
//	})
//}
