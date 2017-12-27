package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"io/ioutil"
	"runtime"
	"strings"
)

func init() {
	runtime.LockOSThread()
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func main() {
	// +Setup GLFW
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape {
			w.SetShouldClose(true)
		}
	})
	window.MakeContextCurrent()
	glfw.SwapInterval(0)
	// -Setup GLFW

	// +Setup GL
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var major int32
	var minor int32
	gl.GetIntegerv(gl.MAJOR_VERSION, &major)
	gl.GetIntegerv(gl.MINOR_VERSION, &minor)
	fmt.Printf("GL %d.%d\n", major, minor)

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LESS)
	gl.Viewport(0, 0, 640, 480)
	gl.ClearColor(0.5, 0.5, 1.0, 1.0)
	// -Setup GL

	// +Setup shaders
	vertSource, err := ioutil.ReadFile("simple.vert")
	if err != nil {
		panic(err)
	}
	fragSource, err := ioutil.ReadFile("simple.frag")
	if err != nil {
		panic(err)
	}
	program, err := newProgram(string(vertSource), string(fragSource))
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	{
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), 640.0/480.0, 1.0, 100.0)
		fmt.Printf("%v\n", projection)
		projectionUniform := gl.GetUniformLocation(program, gl.Str("pMatrix\x00"))
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		model := mgl32.Translate3D(-10.0, -10.0, -10.0)
		modelUniform := gl.GetUniformLocation(program, gl.Str("mvMatrix\x00"))
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])
	}
	// -Setup shaders

	// +Setup geom
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var verts = []float32{}
	var colors = []float32{}
	var indices = []int32{}
	{
		width := 100
		height := 100
		for y := 0; y < height+1; y++ {
			for x := 0; x < width+1; x++ {
				fx := float32(x)
				fy := float32(y)
				verts = append(verts, []float32{
					fx, fy, -20.0,
				}...)
				shade := float32(x^y) / 40.0
				colors = append(colors, []float32{
					shade, shade, shade,
				}...)
			}
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				bl := int32(((width + 1) * y) + x)
				br := bl + 1
				tl := int32(((width + 1) * (y + 1)) + x)
				tr := tl + 1
				indices = append(indices, []int32{
					bl, tr, tl,
					bl, br, tr,
				}...)
			}
		}

		fmt.Printf("%v\n", verts)
		fmt.Printf("%v\n", indices)

		var vbo uint32
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(verts), gl.STATIC_DRAW)

		vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertex\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		var vbo2 uint32
		gl.GenBuffers(1, &vbo2)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, vbo2)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

		var vbo3 uint32
		gl.GenBuffers(1, &vbo3)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo3)
		gl.BufferData(gl.ARRAY_BUFFER, len(colors)*4, gl.Ptr(colors), gl.STATIC_DRAW)

		colorAttrib := uint32(gl.GetAttribLocation(program, gl.Str("color\x00")))
		gl.EnableVertexAttribArray(colorAttrib)
		gl.VertexAttribPointer(colorAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}
	// -Setup geom

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// +Draw geom
		gl.BindVertexArray(vao)
		//gl.DrawArrays(gl.LINES, 0, 6)
		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))
		// -Draw geom

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
