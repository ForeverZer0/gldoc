# GLdoc

GLdoc hosts a local server that returns OpenGL documentation parsed from the [OpenGL-Refpages](https://github.com/KhronosGroup/OpenGL-Refpages) in a simplified JSON interface, which can then be read and used from any language using only basic HTTP requests.

## Purpose

Generating bindings for OpenGL using the [OpenGL registry](https://github.com/KhronosGroup/OpenGL-Registry) is pretty standard boilerplate for many projects. While generating the API bindings is typically trivial, including documentation for said bindings is often less so, as the sources are primarily intended for generating a static website in HTML, and less so for code documentation. As someone who likes well-documented code and is an avid enjoyer of inline hints, this tool was created to solve that (admittedly niche) problem.

## Overview

Before getting into any details, a simple example will make what it does evident...

Start the server:
```bash
gldoc --api=gl --version=3.3
```
Now make a request to the server (using default address):
```bash
curl --silent localhost:8888/glBufferData | jq
```

The resulting reply:
```json
{
  "name": "glBufferData",
  "desc": "creates and initializes a buffer object's data store",
  "args": {
    "target": "Specifies the target to which the buffer object is bound for glBufferData, which must be one of the buffer binding targets in the following table:",
    "size": "Specifies the size in bytes of the buffer object's new data store.",
    "data": "Specifies a pointer to data that will be copied into the data store for initialization, or NULL if no data is to be copied.",
    "usage": "Specifies the expected usage pattern of the data store. The symbolic constant must be GL_STREAM_DRAW, GL_STREAM_READ, GL_STREAM_COPY, GL_STATIC_DRAW, GL_STATIC_READ, GL_STATIC_COPY, GL_DYNAMIC_DRAW, GL_DYNAMIC_READ, or GL_DYNAMIC_COPY."
  },
  "seealso": [
    "glBindBuffer",
    "glBufferSubData",
    "glMapBuffer",
    "glUnmapBuffer"
  ],
  "errors": [
    "GL_INVALID_ENUM",
    "GL_INVALID_VALUE",
    "GL_INVALID_OPERATION",
    "GL_OUT_OF_MEMORY"
  ]
}
```
