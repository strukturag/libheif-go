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

import (
	"fmt"
	"image"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	require := require.New(t)
	version := GetVersion()
	require.NotEmpty(version, "Version is missing")
}

type decodeTest struct {
	colorspace Colorspace
	chroma     Chroma
}

func CheckHeifImage(t *testing.T, handle *ImageHandle, thumbnail bool) {
	t.Helper()

	t.Run("properties", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		handle.GetWidth()
		handle.GetHeight()
		handle.HasAlphaChannel()
		handle.HasDepthImage()
		count := handle.GetNumberOfDepthImages()
		ids := handle.GetListOfDepthImageIDs()
		assert.Len(ids, count, "Number of depth image ids mismatched")
		if !thumbnail {
			count = handle.GetNumberOfThumbnails()
			ids := handle.GetListOfThumbnailIDs()
			assert.Len(ids, count, "Number of thumbnail image ids mismatched")
			for _, id := range ids {
				id := id
				t.Run(fmt.Sprintf("thumb-%d", id), func(t *testing.T) {
					t.Parallel()
					if thumb, err := handle.GetThumbnail(id); assert.NoError(err) {
						CheckHeifImage(t, thumb, true)
					}
				})
			}
		}
	})

	decodeTests := []decodeTest{
		{ColorspaceUndefined, ChromaUndefined},
		{ColorspaceYCbCr, Chroma420},
		{ColorspaceYCbCr, Chroma422},
		{ColorspaceYCbCr, Chroma444},
		{ColorspaceRGB, Chroma444},
		{ColorspaceRGB, ChromaInterleavedRGB},
		{ColorspaceRGB, ChromaInterleavedRGBA},
		{ColorspaceRGB, ChromaInterleavedRRGGBB_BE},
		{ColorspaceRGB, ChromaInterleavedRRGGBBAA_BE},
	}
	for _, test := range decodeTests {
		test := test
		t.Run(fmt.Sprintf("%d/%d", test.colorspace, test.chroma), func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			if img, err := handle.DecodeImage(test.colorspace, test.chroma, nil); assert.NoError(err, "Error decoding image") {
				img.GetColorspace()
				img.GetChromaFormat()

				_, err := img.GetImage()
				assert.NoError(err)
			}
		})
	}
}

func CheckHeifFile(t *testing.T, ctx *Context) {
	t.Helper()
	assert := assert.New(t)
	assert.Equal(2, ctx.GetNumberOfTopLevelImages(), "Number of top level images mismatched")
	assert.Len(ctx.GetListOfTopLevelImageIDs(), 2, "Number of top level image ids mismatched")
	_, err := ctx.GetPrimaryImageID()
	assert.NoError(err, "Error getting primary image id")
	if handle, err := ctx.GetPrimaryImageHandle(); assert.NoError(err, "Error getting primary image handler") {
		assert.True(handle.IsPrimaryImage(), "Expected primary image")

		t.Run("primary", func(t *testing.T) {
			t.Parallel()
			CheckHeifImage(t, handle, false)
		})
	}
}

func TestReadFromFile(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	ctx, err := NewContext()
	require.NoError(err, "Can't create context")

	filename := path.Join("testdata", "example.heic")
	require.NoError(ctx.ReadFromFile(filename))

	CheckHeifFile(t, ctx)
}

func TestReadFromMemory(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	ctx, err := NewContext()
	require.NoError(err, "Can't create context")

	filename := path.Join("testdata", "example.heic")
	data, err := os.ReadFile(filename)
	require.NoError(err, "Can't read file")
	require.NoError(ctx.ReadFromMemory(data))

	// Make sure future processing works if "data" is GC'd
	data = nil // nolint
	runtime.GC()

	CheckHeifFile(t, ctx)
}

func TestReadImage(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	require := require.New(t)
	filename := path.Join("testdata", "example.heic")
	fp, err := os.Open(filename)
	require.NoError(err)
	defer fp.Close()

	config, format1, err := image.DecodeConfig(fp)
	require.NoError(err)
	assert.Equal("heif", format1)
	_, err = fp.Seek(0, 0)
	require.NoError(err)

	img, format2, err := image.Decode(fp)
	require.NoError(err)
	assert.Equal("heif", format2)

	r := img.Bounds()
	if config.Width != (r.Max.X-r.Min.X) || config.Height != (r.Max.Y-r.Min.Y) {
		fmt.Printf("Image size %+v does not match config %+v\n", r, config)
	}
}
