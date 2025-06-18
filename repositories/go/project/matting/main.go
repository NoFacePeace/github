package main

import (
	"fmt"
	"image/color"

	"gocv.io/x/gocv"
)

func main() {
	// 读取输入图像
	img := gocv.IMRead("images/jk.jpg", gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error: Could not read image")
		return
	}
	defer img.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	edges := gocv.NewMat()
	defer edges.Close()
	// gocv.Canny(gray, &edges, 50, 150)

	gocv.Threshold(gray, &edges, 100, 255, gocv.ThresholdBinary)

	contours := gocv.FindContours(edges, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	gocv.DrawContours(&img, contours, -1, color.RGBA{255, 0, 0, 255}, 2)
	window := gocv.NewWindow("jk")
	window.IMShow(img)
	window.WaitKey(0)
	window.Close()

}
