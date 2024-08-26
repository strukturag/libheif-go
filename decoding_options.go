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
	"errors"
	"runtime"
)

// DecodingOptions contain options that are used for decoding.
type DecodingOptions struct {
	options *C.struct_heif_decoding_options
}

func freeHeifDecodingOptions(options *DecodingOptions) {
	C.heif_decoding_options_free(options.options)
	options.options = nil
}

// NewDecodingOptions creates new decoding options.
func NewDecodingOptions() (*DecodingOptions, error) {
	if err := checkLibraryVersion(); err != nil {
		return nil, err
	}

	options := &DecodingOptions{
		options: C.heif_decoding_options_alloc(),
	}
	if options.options == nil {
		return nil, errors.New("Could not allocate decoding options")
	}

	runtime.SetFinalizer(options, freeHeifDecodingOptions)
	options.options.version = 5
	return options, nil
}
