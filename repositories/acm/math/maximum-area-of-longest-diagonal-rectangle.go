package math

func areaOfMaxDiagonal(dimensions [][]int) int {
	height := 0
	width := 0
	for _, dimension := range dimensions {
		w, h := dimension[0], dimension[1]
		if w*w+h*h < height*height+width*width {
			continue
		}
		if w*w+h*h > height*height+width*width {
			height = h
			width = w
			continue
		}
		if w*h > height*width {
			height = h
			width = w
		}
	}
	return height * width
}
