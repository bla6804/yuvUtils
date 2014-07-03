package main

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
)

type MyYCbCr image.YCbCr

func (p *MyYCbCr) Read(r io.Reader) (int, error) {
	bytesRead := 0

	n, err := r.Read(p.Y)
	bytesRead += n
	if err != nil {
		return bytesRead, err
	}
	if n != len(p.Y) {
		return bytesRead, errors.New("Wrong size")
	}

	n, err = r.Read(p.Cb)
	bytesRead += n
	if err != nil {
		return bytesRead, err
	}
	if n != len(p.Cb) {
		return bytesRead, errors.New("Wrong size")
	}

	n, err = r.Read(p.Cr)
	bytesRead += n
	if err != nil {
		return bytesRead, err
	}
	if n != len(p.Cr) {
		return bytesRead, errors.New("Wrong size")
	}

	return bytesRead, nil
}

func (p *MyYCbCr) Write(w io.Writer) (int, error) {
	bytesWritten := 0

	n, err := w.Write(p.Y)
	bytesWritten += n
	if err != nil {
		return bytesWritten, err
	}
	if n != len(p.Y) {
		return bytesWritten, errors.New("Wrong size")
	}

	n, err = w.Write(p.Cb)
	bytesWritten += n
	if err != nil {
		return bytesWritten, err
	}
	if n != len(p.Cb) {
		return bytesWritten, errors.New("Wrong size")
	}

	n, err = w.Write(p.Cr)
	bytesWritten += n
	if err != nil {
		return bytesWritten, err
	}
	if n != len(p.Cr) {
		return bytesWritten, errors.New("Wrong size")
	}

	return bytesWritten, nil
}

func Merge(frame1, frame2, frameMerged *MyYCbCr) error {
	if frame1.Rect.Dy() != frame2.Rect.Dy() ||
		frame1.Rect.Dy() != frameMerged.Rect.Dy() {
		return errors.New("Wrong height")
	}
	if frame1.Rect.Dx()+frame2.Rect.Dx() != frameMerged.Rect.Dx() {
		return errors.New("Wrong width")
	}

	// frame1 => left side
	for line := 0; line < frame1.Rect.Dy(); line++ {
		dStart := line * frameMerged.YStride
		dStop := dStart + frame1.YStride
		sStart := line * frame1.YStride
		sStop := sStart + frame1.YStride
		copy(frameMerged.Y[dStart:dStop], frame1.Y[sStart:sStop])
	}
	for line := 0; line < len(frame1.Cb)/frame1.CStride; line++ {
		dStart := line * frameMerged.CStride
		dStop := dStart + frame1.CStride
		sStart := line * frame1.CStride
		sStop := sStart + frame1.CStride
		copy(frameMerged.Cb[dStart:dStop], frame1.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame1.Cr[sStart:sStop])
	}

	// frame 2 => right side
	for line := 0; line < frame1.Rect.Dy(); line++ {
		dStart := line*frameMerged.YStride + frame1.YStride
		dStop := dStart + frame2.YStride
		sStart := line * frame2.YStride
		sStop := sStart + frame2.YStride
		copy(frameMerged.Y[dStart:dStop], frame2.Y[sStart:sStop])
	}
	for line := 0; line < len(frame2.Cb)/frame2.CStride; line++ {
		dStart := line*frameMerged.CStride + frame1.CStride
		dStop := dStart + frame2.CStride
		sStart := line * frame2.CStride
		sStop := sStart + frame2.CStride
		copy(frameMerged.Cb[dStart:dStop], frame2.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame2.Cr[sStart:sStop])
	}

	return nil
}

func main() {
	fileName1 := "tag248_426x240.yuv"
	fileName2 := "tag133_426x240.yuv"
	width := 426
	height := 240

	file1, err := os.Open(fileName1)
	if err != nil {
		log.Fatal(err)
	}
	defer file1.Close()

	file2, err := os.Open(fileName2)
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	fileOut, err := os.Create("bla_852x240.yuv")
	if err != nil {
		log.Fatal(err)
	}
	defer fileOut.Close()

	frame1 := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frame2 := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frameMerged := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, 2*width, height),
		image.YCbCrSubsampleRatio420))
	var i int
	for i = 0; ; i++ {
		_, err := frame1.Read(file1)
		if err != nil {
			break
		}
		_, err = frame2.Read(file2)
		if err != nil {
			break
		}
		err = Merge(frame1, frame2, frameMerged)
		frameMerged.Write(fileOut)

	}

	fmt.Println("frames:", i)
}
