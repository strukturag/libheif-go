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
	"runtime"
)

// Encoder contains a libheif encoder object.
type Encoder struct {
	encoder *C.struct_heif_encoder
	id      string
	name    string
}

func freeHeifEncoder(enc *Encoder) {
	C.heif_encoder_release(enc.encoder)
	enc.encoder = nil
}

// ID returns the id of the encoder.
func (e *Encoder) ID() string {
	return e.id
}

// Name returns the name of the encoder.
func (e *Encoder) Name() string {
	return e.name
}

// SetQuality sets the quality level for the encoder.
func (e *Encoder) SetQuality(q int) error {
	defer runtime.KeepAlive(e)

	err := C.heif_encoder_set_lossy_quality(e.encoder, C.int(q))
	return convertHeifError(err)
}

// SetLossless enables or disables the lossless encoding mode.
func (e *Encoder) SetLossless(l LosslessMode) error {
	defer runtime.KeepAlive(e)

	err := C.heif_encoder_set_lossless(e.encoder, C.int(l))
	return convertHeifError(err)
}

// SetLoggingLevel sets the logging level to use for encoding.
func (e *Encoder) SetLoggingLevel(l LoggingLevel) error {
	defer runtime.KeepAlive(e)

	err := C.heif_encoder_set_logging_level(e.encoder, C.int(l))
	return convertHeifError(err)
}
