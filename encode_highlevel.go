/*
 * Go interface to libheif
 *
 * Copyright (c) 2018-2024 struktur AG, Joachim Bauch <bauch@struktur.de>
 *
 * libheif is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as
 * published by the Free Software Foundation, either version 3 of
 * the License, or (at your option) any later version.
 *
 * libheif is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with libheif.  If not, see <http://www.gnu.org/licenses/>.
 */

package libheif

// #cgo pkg-config: libheif
// #include <stdlib.h>
// #include <string.h>
// #include <libheif/heif.h>
import "C"

import (
	"fmt"
	"image"
	"runtime"
)

func imageFromRGBA(i *image.RGBA) (*Image, error) {
	min := i.Bounds().Min
	max := i.Bounds().Max
	w := max.X - min.X
	h := max.Y - min.Y

	out, err := NewImage(w, h, ColorspaceRGB, ChromaInterleavedRGBA)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %v", err)
	}

	p, err := out.NewPlane(ChannelInterleaved, w, h, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to add plane: %v", err)
	}
	p.setData([]byte(i.Pix), w*4)

	return out, nil
}

func imageFromNRGBA(i *image.NRGBA) (*Image, error) {
	min := i.Bounds().Min
	max := i.Bounds().Max
	w := max.X - min.X
	h := max.Y - min.Y

	out, err := NewImage(w, h, ColorspaceRGB, ChromaInterleavedRGBA)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %v", err)
	}

	p, err := out.NewPlane(ChannelInterleaved, w, h, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to add plane: %v", err)
	}
	p.setData([]byte(i.Pix), w*4)

	return out, nil
}

func imageFromRGBA64(i *image.RGBA64) (*Image, error) {
	min := i.Bounds().Min
	max := i.Bounds().Max
	w := max.X - min.X
	h := max.Y - min.Y

	out, err := NewImage(w, h, ColorspaceRGB, ChromaInterleavedRRGGBBAA_BE)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %v", err)
	}

	p, err := out.NewPlane(ChannelInterleaved, w, h, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to add plane: %v", err)
	}

	pix := make([]byte, w*h*8)
	read_pos := 0
	write_pos := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := (uint16(i.Pix[read_pos]) << 8) | uint16(i.Pix[read_pos+1])
			r = r >> 6
			pix[write_pos] = byte(r >> 8)
			pix[write_pos+1] = byte(r & 0xff)
			read_pos += 2
			g := (uint16(i.Pix[read_pos]) << 8) | uint16(i.Pix[read_pos+1])
			g = g >> 6
			pix[write_pos+2] = byte(g >> 8)
			pix[write_pos+3] = byte(g & 0xff)
			read_pos += 2
			b := (uint16(i.Pix[read_pos]) << 8) | uint16(i.Pix[read_pos+1])
			b = b >> 6
			pix[write_pos+4] = byte(b >> 8)
			pix[write_pos+5] = byte(b & 0xff)
			read_pos += 2
			a := (uint16(i.Pix[read_pos]) << 8) | uint16(i.Pix[read_pos+1])
			a = a >> 6
			pix[write_pos+6] = byte(a >> 8)
			pix[write_pos+7] = byte(a & 0xff)
			pix[write_pos+6] = byte(a >> 8)
			pix[write_pos+7] = byte(a & 0xff)
			read_pos += 2
			write_pos += 8
		}
	}
	p.setData(pix, w*8)

	return out, nil
}

func imageFromGray(i *image.Gray) (*Image, error) {
	min := i.Bounds().Min
	max := i.Bounds().Max
	w := max.X - min.X
	h := max.Y - min.Y

	out, err := NewImage(w, h, ColorspaceYCbCr, ChromaMonochrome)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %v", err)
	}

	const depth = 8
	pY, err := out.NewPlane(ChannelY, w, h, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to add Y plane: %v", err)
	}
	pY.setData([]byte(i.Pix), i.Stride)

	return out, nil
}

func imageFromYCbCr(i *image.YCbCr) (*Image, error) {
	min := i.Bounds().Min
	max := i.Bounds().Max
	w := max.X - min.X
	h := max.Y - min.Y

	var cm Chroma
	switch sr := i.SubsampleRatio; sr {
	case image.YCbCrSubsampleRatio420:
		cm = Chroma420
	case image.YCbCrSubsampleRatio444:
		cm = Chroma444
	default:
		return nil, fmt.Errorf("unsupported subsample ratio: %s", sr.String())
	}

	out, err := NewImage(w, h, ColorspaceYCbCr, cm)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %v", err)
	}

	const depth = 8
	pY, err := out.NewPlane(ChannelY, w, h, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to add Y plane: %v", err)
	}
	pY.setData([]byte(i.Y), i.YStride)

	switch cm {
	case Chroma420:
		halfW, halfH := (w+1)/2, (h+1)/2
		pCb, err := out.NewPlane(ChannelCb, halfW, halfH, depth)
		if err != nil {
			return nil, fmt.Errorf("failed to add Cb plane: %v", err)
		}
		pCb.setData([]byte(i.Cb), i.CStride)
		pCr, err := out.NewPlane(ChannelCr, halfW, halfH, depth)
		if err != nil {
			return nil, fmt.Errorf("failed to add Cr plane: %v", err)
		}
		pCr.setData([]byte(i.Cr), i.CStride)
	case Chroma444:
		pCb, err := out.NewPlane(ChannelCb, w, h, depth)
		if err != nil {
			return nil, fmt.Errorf("failed to add Cb plane: %v", err)
		}
		pCb.setData([]byte(i.Cb), i.CStride)
		pCr, err := out.NewPlane(ChannelCr, w, h, depth)
		if err != nil {
			return nil, fmt.Errorf("failed to add Cr plane: %v", err)
		}
		pCr.setData([]byte(i.Cr), i.CStride)
	}

	return out, nil
}

// EncodeFromImage is a high-level function to encode a Go Image to a new Context.
func EncodeFromImage(img image.Image, compression CompressionFormat, quality int, lossless LosslessMode, logging LoggingLevel) (*Context, error) {
	if err := checkLibraryVersion(); err != nil {
		return nil, err
	}

	var out *Image

	switch i := img.(type) {
	default:
		return nil, fmt.Errorf("unsupported image type: %T", i)
	case *image.RGBA:
		tmp, err := imageFromRGBA(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		out = tmp
	case *image.NRGBA:
		tmp, err := imageFromNRGBA(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		out = tmp
	case *image.RGBA64:
		tmp, err := imageFromRGBA64(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		out = tmp
	case *image.Gray:
		tmp, err := imageFromGray(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		out = tmp
	case *image.YCbCr:
		tmp, err := imageFromYCbCr(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		out = tmp
	}

	ctx, err := NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create HEIF context: %v", err)
	}

	enc, err := ctx.NewEncoder(compression)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %v", err)
	}

	if err := enc.SetQuality(quality); err != nil {
		return nil, fmt.Errorf("failed to set quality: %v", err)
	}
	if err := enc.SetLossless(lossless); err != nil {
		return nil, fmt.Errorf("failed to set lossless mode: %v", err)
	}
	if err := enc.SetLoggingLevel(logging); err != nil {
		return nil, fmt.Errorf("failed to set logging level: %v", err)
	}

	encOpts, err := NewEncodingOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding options: %v", err)
	}

	defer runtime.KeepAlive(ctx)
	defer runtime.KeepAlive(out)
	defer runtime.KeepAlive(enc)
	defer runtime.KeepAlive(encOpts)

	var handle ImageHandle
	err2 := C.heif_context_encode_image(ctx.context, out.image, enc.encoder, encOpts.options, &handle.handle)
	if err := convertHeifError(err2); err != nil {
		return nil, fmt.Errorf("failed to encode image: %v", err)
	}

	runtime.SetFinalizer(&handle, freeHeifImageHandle)
	return ctx, nil
}
