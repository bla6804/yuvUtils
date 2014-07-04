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

func DiffFrames(frame1, frame2, frameDiff *MyYCbCr, factor int) error {
	if frame1.Rect.Dy() != frame2.Rect.Dy() ||
		frame1.Rect.Dy() != frameDiff.Rect.Dy() {
		return errors.New("Height differs")
	}
	if frame1.Rect.Dx() != frame2.Rect.Dx() ||
		frame1.Rect.Dx() != frameDiff.Rect.Dx() {
		return errors.New("Width differs")
	}

	for i := range frame1.Y {
		tmp := (int(frame1.Y[i])-int(frame2.Y[i]))*factor + 128
		if tmp < 0 {
			tmp = 0
		}
		if tmp > 255 {
			tmp = 255
		}
		frameDiff.Y[i] = uint8(tmp)
	}

	for i := range frame1.Cb {
		tmp := (int(frame1.Cb[i])-int(frame2.Cb[i]))*factor + 128
		if tmp < 0 {
			tmp = 0
		}
		if tmp > 255 {
			tmp = 255
		}
		frameDiff.Cb[i] = uint8(tmp)
	}

	for i := range frame1.Cr {
		tmp := (int(frame1.Cr[i])-int(frame2.Cr[i]))*factor + 128
		if tmp < 0 {
			tmp = 0
		}
		if tmp > 255 {
			tmp = 255
		}
		frameDiff.Cr[i] = uint8(tmp)
	}

	return nil
}

func Merge2(frame1, frame2, frameMerged *MyYCbCr) error {
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

func Merge4(frame1, frame2, frame3, frame4, frameMerged *MyYCbCr) error {
	if 2*frame1.Rect.Dy() != frameMerged.Rect.Dy() ||
		2*frame2.Rect.Dy() != frameMerged.Rect.Dy() ||
		2*frame3.Rect.Dy() != frameMerged.Rect.Dy() ||
		2*frame4.Rect.Dy() != frameMerged.Rect.Dy() {
		return errors.New("Wrong height")
	}
	if 2*frame1.Rect.Dx() != frameMerged.Rect.Dx() ||
		2*frame2.Rect.Dx() != frameMerged.Rect.Dx() ||
		2*frame3.Rect.Dx() != frameMerged.Rect.Dx() ||
		2*frame4.Rect.Dx() != frameMerged.Rect.Dx() {
		return errors.New("Wrong width")
	}

	cHeight1 := len(frame1.Cb) / frame1.CStride
	cHeight2 := len(frame2.Cb) / frame2.CStride
	cHeight3 := len(frame3.Cb) / frame3.CStride
	cHeight4 := len(frame4.Cb) / frame4.CStride

	// frame1 => top left
	for line := 0; line < frame1.Rect.Dy(); line++ {
		dStart := line * frameMerged.YStride
		dStop := dStart + frame1.YStride
		sStart := line * frame1.YStride
		sStop := sStart + frame1.YStride
		copy(frameMerged.Y[dStart:dStop], frame1.Y[sStart:sStop])
	}
	for line := 0; line < cHeight1; line++ {
		dStart := line * frameMerged.CStride
		dStop := dStart + frame1.CStride
		sStart := line * frame1.CStride
		sStop := sStart + frame1.CStride
		copy(frameMerged.Cb[dStart:dStop], frame1.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame1.Cr[sStart:sStop])
	}

	// frame 2 => top right
	for line := 0; line < frame1.Rect.Dy(); line++ {
		dStart := line*frameMerged.YStride + frame1.YStride
		dStop := dStart + frame2.YStride
		sStart := line * frame2.YStride
		sStop := sStart + frame2.YStride
		copy(frameMerged.Y[dStart:dStop], frame2.Y[sStart:sStop])
	}
	for line := 0; line < cHeight2; line++ {
		dStart := line*frameMerged.CStride + frame1.CStride
		dStop := dStart + frame2.CStride
		sStart := line * frame2.CStride
		sStop := sStart + frame2.CStride
		copy(frameMerged.Cb[dStart:dStop], frame2.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame2.Cr[sStart:sStop])
	}

	// frame3 => bottom left
	for line := 0; line < frame3.Rect.Dy(); line++ {
		dStart := (line + frame1.Rect.Dy()) * frameMerged.YStride
		dStop := dStart + frame3.YStride
		sStart := line * frame3.YStride
		sStop := sStart + frame3.YStride
		copy(frameMerged.Y[dStart:dStop], frame3.Y[sStart:sStop])
	}
	for line := 0; line < cHeight3; line++ {
		dStart := (line + cHeight1) * frameMerged.CStride
		dStop := dStart + frame3.CStride
		sStart := line * frame3.CStride
		sStop := sStart + frame3.CStride
		copy(frameMerged.Cb[dStart:dStop], frame3.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame3.Cr[sStart:sStop])
	}

	// frame 4 => bottom right
	for line := 0; line < frame4.Rect.Dy(); line++ {
		dStart := (line+frame2.Rect.Dy())*frameMerged.YStride + frame3.YStride
		dStop := dStart + frame4.YStride
		sStart := line * frame4.YStride
		sStop := sStart + frame4.YStride
		copy(frameMerged.Y[dStart:dStop], frame4.Y[sStart:sStop])
	}
	for line := 0; line < cHeight4; line++ {
		dStart := (line+cHeight2)*frameMerged.CStride + frame3.CStride
		dStop := dStart + frame4.CStride
		sStart := line * frame4.CStride
		sStop := sStart + frame4.CStride
		copy(frameMerged.Cb[dStart:dStop], frame4.Cb[sStart:sStop])
		copy(frameMerged.Cr[dStart:dStop], frame4.Cr[sStart:sStop])
	}

	return nil
}

func main() {
	src1Filename := "tag248_426x240.yuv"
	src2Filename := "tag133_426x240.yuv"
	destFilename := "diff_426x240.yuv"
	width := 426
	height := 240

	src1, err := os.Open(src1Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer src1.Close()

	src2, err := os.Open(src2Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer src2.Close()

	dest, err := os.Create(destFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer dest.Close()

	frame1 := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frame2 := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frameDiff := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frameDiff2 := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, width, height),
		image.YCbCrSubsampleRatio420))
	frameMerged := (*MyYCbCr)(image.NewYCbCr(
		image.Rect(0, 0, 2*width, 2*height),
		image.YCbCrSubsampleRatio420))
	var i int
	for i = 0; ; i++ {
		_, err := frame1.Read(src1)
		if err != nil {
			break
		}
		_, err = frame2.Read(src2)
		if err != nil {
			break
		}

		err = DiffFrames(frame1, frame2, frameDiff, 1)
		if err != nil {
			log.Fatal(err)
		}
		err = DiffFrames(frame1, frame2, frameDiff2, 10)
		if err != nil {
			log.Fatal(err)
		}
		err = Merge4(frame1, frame2, frameDiff, frameDiff2, frameMerged)
		if err != nil {
			log.Fatal(err)
		}
		frameMerged.Write(dest)
	}

	fmt.Println("frames:", i)
}
