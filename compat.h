/*
 * Compatibility shim for libheif header changes across versions.
 *
 * Older libheif releases (<= 1.17) declare enums such as "heif_chroma" as
 * plain, un-typedef'd "enum heif_xxx" tags, while newer releases (>= 1.19)
 * declare them via "typedef enum heif_xxx { ... } heif_xxx;". Both forms are
 * freely interchangeable in C, but cgo treats "C.enum_heif_xxx" and
 * "C.heif_xxx" as distinct, incompatible Go types, and only one of the two
 * spellings is guaranteed to exist depending on which libheif header is
 * installed. Passing a plain integer across the cgo boundary and letting C
 * perform the (always legal) implicit conversion to the enum type sidesteps
 * the ambiguity entirely, so every call into libheif that takes one of these
 * enums goes through a thin wrapper defined here instead of calling the
 * libheif function directly from Go.
 */

#ifndef LIBHEIF_GO_COMPAT_H
#define LIBHEIF_GO_COMPAT_H

#include <libheif/heif.h>

static inline int go_heif_have_decoder_for_format(uint32_t format)
{
  return heif_have_decoder_for_format((enum heif_compression_format)format);
}

static inline int go_heif_have_encoder_for_format(uint32_t format)
{
  return heif_have_encoder_for_format((enum heif_compression_format)format);
}

static inline int go_heif_context_get_encoder_descriptors(struct heif_context* ctx,
                                                            uint32_t format,
                                                            const char* name,
                                                            const struct heif_encoder_descriptor** out_descriptors,
                                                            int count)
{
  return heif_context_get_encoder_descriptors(ctx, (enum heif_compression_format)format, name, out_descriptors, count);
}

static inline void go_heif_set_preferred_chroma_downsampling_algorithm(struct heif_decoding_options* options,
                                                                        uint32_t algorithm)
{
  options->color_conversion_options.preferred_chroma_downsampling_algorithm =
      (enum heif_chroma_downsampling_algorithm)algorithm;
}

static inline uint32_t go_heif_get_preferred_chroma_downsampling_algorithm(struct heif_decoding_options* options)
{
  return (uint32_t)options->color_conversion_options.preferred_chroma_downsampling_algorithm;
}

static inline void go_heif_set_preferred_chroma_upsampling_algorithm(struct heif_decoding_options* options,
                                                                      uint32_t algorithm)
{
  options->color_conversion_options.preferred_chroma_upsampling_algorithm =
      (enum heif_chroma_upsampling_algorithm)algorithm;
}

static inline uint32_t go_heif_get_preferred_chroma_upsampling_algorithm(struct heif_decoding_options* options)
{
  return (uint32_t)options->color_conversion_options.preferred_chroma_upsampling_algorithm;
}

static inline struct heif_error go_heif_decode_image(const struct heif_image_handle* in_handle,
                                                      struct heif_image** out_img,
                                                      uint32_t colorspace,
                                                      uint32_t chroma,
                                                      const struct heif_decoding_options* options)
{
  return heif_decode_image(in_handle, out_img, (enum heif_colorspace)colorspace, (enum heif_chroma)chroma, options);
}

static inline struct heif_error go_heif_image_create(int width, int height,
                                                      uint32_t colorspace,
                                                      uint32_t chroma,
                                                      struct heif_image** out_image)
{
  return heif_image_create(width, height, (enum heif_colorspace)colorspace, (enum heif_chroma)chroma, out_image);
}

static inline int go_heif_image_get_width(const struct heif_image* img, uint32_t channel)
{
  return heif_image_get_width(img, (enum heif_channel)channel);
}

static inline int go_heif_image_get_height(const struct heif_image* img, uint32_t channel)
{
  return heif_image_get_height(img, (enum heif_channel)channel);
}

static inline int go_heif_image_get_bits_per_pixel(const struct heif_image* img, uint32_t channel)
{
  return heif_image_get_bits_per_pixel(img, (enum heif_channel)channel);
}

static inline int go_heif_image_get_bits_per_pixel_range(const struct heif_image* img, uint32_t channel)
{
  return heif_image_get_bits_per_pixel_range(img, (enum heif_channel)channel);
}

static inline uint8_t* go_heif_image_get_plane(struct heif_image* img, uint32_t channel, int* out_stride)
{
  return heif_image_get_plane(img, (enum heif_channel)channel, out_stride);
}

static inline struct heif_error go_heif_image_add_plane(struct heif_image* image, uint32_t channel, int width, int height, int bit_depth)
{
  return heif_image_add_plane(image, (enum heif_channel)channel, width, height, bit_depth);
}

#endif
