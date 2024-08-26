#!/usr/bin/python3
"""
Go interface to libheif

Copyright (c) 2024 struktur AG, Joachim Bauch <bauch@struktur.de>

libheif is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as
published by the Free Software Foundation, either version 3 of
the License, or (at your option) any later version.

libheif is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with libheif.  If not, see <http://www.gnu.org/licenses/>.
"""
import io
import re
import sys

try:
  ARCH = getattr(sys, "implementation", sys)._multiarch
except AttributeError:
  # This is a non-multiarch aware Python.  Fallback to the old way.
  ARCH = sys.platform

HEADER_FILE="/usr/include/%s/libheif/heif.h" % (ARCH)
VERSION_HEADER_FILE="/usr/include/%s/libheif/heif_version.h" % (ARCH)

HEADER="""/*
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

const build_version = %s
"""

def findHeifEnumValues(enum: str, data: str):
  m = re.search(r'\nenum %s\s*{[^}]*};\n' % (enum), data)
  if m is None:
    raise TypeError('Enum %s not found' % (enum))

  start = m.start()
  end = m.end()
  last_end = start + 6 + len(enum)
  # skip whitespaces
  while data[last_end] in ' \n\r\t':
    last_end += 1
  # skip curly bracket
  last_end += 1
  # skip whitespaces
  while data[last_end] in ' \n\r\t':
    last_end += 1

  if enum == 'heif_error_code':
    enum = 'heif_error'
  elif enum == 'heif_suberror_code':
    enum = 'heif_suberror'
  elif enum == 'heif_compression_format':
    enum = 'heif_compression'
  elif enum == 'heif_chroma_downsampling_algorithm':
    enum = 'heif_chroma_downsampling'
  elif enum == 'heif_chroma_upsampling_algorithm':
    enum = 'heif_chroma_upsampling'

  entries = []
  # Find enum entry with optional comment in the same line.
  regex = re.compile(r'^[ \t]*(%s_[^=\s]+)\s*=\s*\d+[ \t]*(?:,?[ \t]*\n|,?[ \t]*//\s*([^\n]+))' % (enum), re.MULTILINE)
  while start < end:
    m = regex.search(data, start, end)
    if m is None:
      break

    entry, comment = m.groups()
    if comment is None:
      # Check for comment before enum entry.
      commentEnd = m.start()
      if data[commentEnd-1] in ' \n\r\t':
        commentEnd -= 1

      lines = data[last_end:commentEnd].split('\n')
      comments = []
      while lines:
        lastLine = lines.pop().strip()
        if not lastLine or lastLine[:2] != '//':
          break

        comments.insert(0, lastLine[2:].strip())

      if comments:
        comment = '\n'.join(['\t// ' + x for x in comments])
    else:
      comment = '\t// ' + comment.strip()

    entries.append((entry, comment))
    start = m.end()
    last_end = m.end()

  return entries

GET_UPPERCASE = re.compile('[A-Z]').findall

def isUpperWord(s: str) -> bool:
  count = len(GET_UPPERCASE(s))
  return count > 1

def capitalize(s: str) -> str:
  if not s:
    return s

  s = s[0].upper() + s[1:]
  s = s.replace('Av1', 'AV1')
  return s

def camelCase(s: str) -> str:
  if s[:5] == 'heif_':
    s = s[5:]
  words = s.split('_')
  last = words[-1]
  last2 = len(words) > 2 and words[-2] or ''
  if isUpperWord(last) and isUpperWord(last2):
    words = [capitalize(x) for x in words]
    words = words[:-2] + [last2+'_'+last]
  elif isUpperWord(last):
    words = [capitalize(x) for x in words]
    words[-1] = last
  else:
    words = [capitalize(x) for x in words]
  if len(words) >= 2 and words[0] == words[-1]:
    words = words[:-1]
  if words[-1] == 'Ok':
    words[-1] = 'OK'
  return ''.join(words)

def outputDefines(enum, data, out):
  typeName = camelCase(enum)
  print("type %s C.enum_%s\n" % (typeName, enum), file=out)
  entries = findHeifEnumValues(enum, data)
  if entries:
    print("const (", file=out)
    keys = [camelCase(x[0]) for x in entries]
    longest = sorted(map(len, keys), reverse=True)
    keys = [x.ljust(longest[0]) for x in keys]

    for i, (entry, comment) in enumerate(entries):
      if comment:
        print(comment, file=out)
      print("\t%s %s = C.%s" % (keys[i], typeName, entry), file=out)
    print(")\n", file=out)

def main():
  if len(sys.argv) != 2:
    print('USAGE: %s <filename.go>' % (sys.argv[0]), file=sys.stderr)
    sys.exit(1)

  output_file = sys.argv[1]
  with open(HEADER_FILE, 'r') as fp:
    data = fp.read()

  with open(VERSION_HEADER_FILE, 'r') as fp:
    version_data = fp.read()

  build_version = re.search(r'^#define\s+LIBHEIF_NUMERIC_VERSION\s+\((.+)\)$', version_data, re.MULTILINE).group(1)
  build_version = build_version.strip()
  while '  ' in build_version:
    build_version = build_version.replace('  ', ' ')
  if build_version[-4:] == ' | 0':
    build_version = build_version[:-4]

  out = io.StringIO()
  print(HEADER % (build_version), file=out)

  outputDefines('heif_error_code', data, out)
  outputDefines('heif_suberror_code', data, out)
  outputDefines('heif_compression_format', data, out)
  outputDefines('heif_chroma', data, out)
  outputDefines('heif_colorspace', data, out)
  outputDefines('heif_channel', data, out)
  outputDefines('heif_progress_step', data, out)
  outputDefines('heif_chroma_downsampling_algorithm', data, out)
  outputDefines('heif_chroma_upsampling_algorithm', data, out)

  with open(output_file, 'w') as fp:
    print(out.getvalue().strip(), file=fp)

if __name__ == '__main__':
  main()
