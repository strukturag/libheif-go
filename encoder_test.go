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
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadImage(t *testing.T, filename string) image.Image {
	t.Helper()
	require := require.New(t)
	data, err := os.ReadFile(filename)
	require.NoError(err)

	img, _, err := image.Decode(bytes.NewBuffer(data))
	require.NoError(err)
	return img
}

func TestEncoder(t *testing.T) {
	codecs := map[CompressionFormat]string{
		CompressionHEVC: "heif",
		CompressionAV1:  "avif",
	}

	outdir := "testdata/output"
	err := os.MkdirAll(outdir, 0755)
	require.NotErrorIs(t, err, os.ErrExist)

	for codec, ext := range codecs {
		codec := codec
		ext := ext
		t.Run(ext, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			require.True(HaveEncoderForFormat(codec))

			img := loadImage(t, "testdata/example-1.jpg")
			ctx, err := EncodeFromImage(img, codec, 75, LosslessModeDisabled, LoggingLevelFull)
			require.NoError(err)

			output := path.Join(outdir, "example-1."+ext)
			assert.NoError(ctx.WriteToFile(output))

			var out bytes.Buffer
			assert.NoError(ctx.Write(&out))

			if assert.FileExists(output) {
				if data, err := os.ReadFile(output); assert.NoError(err) {
					assert.Equal(len(data), out.Len())
				}
			}
		})
	}

}
