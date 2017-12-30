package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"io/ioutil"
	"math"
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

func normal(x1, y1, z1, x2, y2, z2, x3, y3, z3 float32) (x, y, z float32) {
	ux := x2 - x1
	uy := y2 - y1
	uz := z2 - z1
	vx := x3 - x1
	vy := y3 - y1
	vz := z3 - z1
	rx := uy*vz - uz*vy
	ry := uz*vx - ux*vz
	rz := ux*vy - uy*vx
	return rx, ry, rz
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
	vertSource[len(vertSource)-1] = 0
	fragSource[len(fragSource)-1] = 0
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

		model := mgl32.Translate3D(-4.0, -3.0, -10.0).Mul4(mgl32.HomogRotate3D(0.4, mgl32.Vec3{1.0, 1.5, 0.0}))
		modelUniform := gl.GetUniformLocation(program, gl.Str("mvMatrix\x00"))
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		normal := model.Inv().Transpose()
		normalUniform := gl.GetUniformLocation(program, gl.Str("normalMatrix\x00"))
		gl.UniformMatrix4fv(normalUniform, 1, false, &normal[0])
	}
	// -Setup shaders

	// +Setup geom
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var verts = []float32{}
	var normals = []float32{}
	var vbo uint32
	var vbo2 uint32
	width := 16
	height := 24
	{
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var ox float32
				var oy float32
				if y%2 == 0 {
					ox = float32(x) * 1.1
					oy = -(float32(y) / 2.0) * 1.1

					verts = append(verts, []float32{
						// Cap
						ox + 0.0, 2.0, oy + 0.5, 1.0,
						ox + 0.5, 2.0, oy - 0.5, 1.0,
						ox - 0.5, 2.0, oy - 0.5, 1.0,

						// +Front right
						ox + 0.0, 0.0, oy + 0.5, 0.0,
						ox + 0.5, 2.0, oy - 0.5, 0.0,
						ox + 0.0, 2.0, oy + 0.5, 0.0,

						ox + 0.0, 0.0, oy + 0.5, 0.0,
						ox + 0.5, 0.0, oy - 0.5, 0.0,
						ox + 0.5, 2.0, oy - 0.5, 0.0,
						// -Front right

						// +Front left
						ox + 0.0, 0.0, oy + 0.5, 0.0,
						ox + 0.0, 2.0, oy + 0.5, 0.0,
						ox - 0.5, 2.0, oy - 0.5, 0.0,

						ox + 0.0, 0.0, oy + 0.5, 0.0,
						ox - 0.5, 2.0, oy - 0.5, 0.0,
						ox - 0.5, 0.0, oy - 0.5, 0.0,
						// -Front left

						// +Back
						ox + 0.5, 2.0, oy - 0.5, 0.0,
						ox + 0.5, 0.0, oy - 0.5, 0.0,
						ox - 0.5, 0.0, oy - 0.5, 0.0,

						ox + 0.5, 2.0, oy - 0.5, 0.0,
						ox - 0.5, 0.0, oy - 0.5, 0.0,
						ox - 0.5, 2.0, oy - 0.5, 0.0,
						// -Back
					}...)

					normals = append(normals, []float32{
						// Cap
						0.0, 1.0, 0.0,
						0.0, 1.0, 0.0,
						0.0, 1.0, 0.0,

						// +Front right
						0.5, 0.0, 0.5,
						0.5, 0.0, 0.5,
						0.5, 0.0, 0.5,

						0.5, 0.0, 0.5,
						0.5, 0.0, 0.5,
						0.5, 0.0, 0.5,
						// -Front right

						// +Front left
						-0.5, 0.0, 0.5,
						-0.5, 0.0, 0.5,
						-0.5, 0.0, 0.5,

						-0.5, 0.0, 0.5,
						-0.5, 0.0, 0.5,
						-0.5, 0.0, 0.5,
						// -Front left

						// +Back
						0.0, 0.0, -1.0,
						0.0, 0.0, -1.0,
						0.0, 0.0, -1.0,

						0.0, 0.0, -1.0,
						0.0, 0.0, -1.0,
						0.0, 0.0, -1.0,
						// -Back
					}...)
				} else {
					ox = (float32(x) + 0.5) * 1.1
					oy = -((float32(y) / 2.0) - 0.5) * 1.1

					verts = append(verts, []float32{
						// Cap
						ox - 0.0, 2.0, oy - 0.5, 1.0,
						ox - 0.5, 2.0, oy + 0.5, 1.0,
						ox + 0.5, 2.0, oy + 0.5, 1.0,

						// +Front right
						ox - 0.0, 0.0, oy - 0.5, 0.0,
						ox - 0.5, 2.0, oy + 0.5, 0.0,
						ox - 0.0, 2.0, oy - 0.5, 0.0,

						ox - 0.0, 0.0, oy - 0.5, 0.0,
						ox - 0.5, 0.0, oy + 0.5, 0.0,
						ox - 0.5, 2.0, oy + 0.5, 0.0,
						// -Front right

						// +Front left
						ox - 0.0, 0.0, oy - 0.5, 0.0,
						ox - 0.0, 2.0, oy - 0.5, 0.0,
						ox + 0.5, 2.0, oy + 0.5, 0.0,

						ox - 0.0, 0.0, oy - 0.5, 0.0,
						ox + 0.5, 2.0, oy + 0.5, 0.0,
						ox + 0.5, 0.0, oy + 0.5, 0.0,
						// -Front left

						// +Back
						ox - 0.5, 2.0, oy + 0.5, 0.0,
						ox - 0.5, 0.0, oy + 0.5, 0.0,
						ox + 0.5, 0.0, oy + 0.5, 0.0,

						ox - 0.5, 2.0, oy + 0.5, 0.0,
						ox + 0.5, 0.0, oy + 0.5, 0.0,
						ox + 0.5, 2.0, oy + 0.5, 0.0,
						// -Back
					}...)
					normals = append(normals, []float32{
						// Cap
						0.0, 1.0, 0.0,
						0.0, 1.0, 0.0,
						0.0, 1.0, 0.0,

						// +Front right
						0.5, 0.0, -0.5,
						0.5, 0.0, -0.5,
						0.5, 0.0, -0.5,

						0.5, 0.0, -0.5,
						0.5, 0.0, -0.5,
						0.5, 0.0, -0.5,
						// -Front right

						// +Front left
						-0.5, 0.0, -0.5,
						-0.5, 0.0, -0.5,
						-0.5, 0.0, -0.5,

						-0.5, 0.0, -0.5,
						-0.5, 0.0, -0.5,
						-0.5, 0.0, -0.5,
						// -Front left

						// +Back
						0.0, 0.0, 1.0,
						0.0, 0.0, 1.0,
						0.0, 0.0, 1.0,

						0.0, 0.0, 1.0,
						0.0, 0.0, 1.0,
						0.0, 0.0, 1.0,
						// -Back
					}...)
				}
			}
		}
		//fmt.Printf("%v\n", verts)

		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(verts)*4, gl.Ptr(verts), gl.DYNAMIC_DRAW)

		vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertex\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		gl.VertexAttribPointer(vertAttrib, 4, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.GenBuffers(1, &vbo2)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
		gl.BufferData(gl.ARRAY_BUFFER, len(normals)*4, gl.Ptr(normals), gl.DYNAMIC_DRAW)

		normalAttrib := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
		gl.EnableVertexAttribArray(normalAttrib)
		gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}
	// -Setup geom

	fmt.Printf("Polys: %d | Vertices: %d | Normals: %d\n", len(verts)/4, (len(verts)/4)*3, len(normals))

	var tick float64 = 0.0
	var frames uint32 = 0
	var tmpVerts = make([]float32, len(verts))
	copy(tmpVerts, verts)
	timeUniform := gl.GetUniformLocation(program, gl.Str("time\x00"))
	for !window.ShouldClose() {
		time := glfw.GetTime()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.Uniform1f(timeUniform, float32(time))

		// +Draw geom
		gl.BindVertexArray(vao)

		// +Update
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				i := ((y * width) + x) * 84

				targetHeight := 2.0 + float32((math.Sin(float64(x)+time)-math.Cos(float64(y)+time))*0.25)

				heights := []float32{targetHeight, targetHeight, targetHeight}
				{
					pos := float64(tmpVerts[i+0] + tmpVerts[i+2])
					heights[0] += float32(math.Sin(pos+time) * 0.25)
				}
				{
					pos := float64(tmpVerts[i+4] + tmpVerts[i+6])
					heights[1] += float32(math.Sin(pos+time) * 0.25)
				}
				{
					pos := float64(tmpVerts[i+8] + tmpVerts[i+10])
					heights[2] += float32(math.Sin(pos+time) * 0.25)
				}

				// +Cap
				tmpVerts[i+1] = heights[0]
				tmpVerts[i+5] = heights[1]
				tmpVerts[i+9] = heights[2]

				{
					t := ((y * width) + x) * 63
					x, y, z := normal(tmpVerts[i+0], tmpVerts[i+1], tmpVerts[i+2], tmpVerts[i+4], tmpVerts[i+5], tmpVerts[i+6], tmpVerts[i+8], tmpVerts[i+9], tmpVerts[i+10])
					normals[t+0] = x
					normals[t+1] = y
					normals[t+2] = z
					normals[t+3] = x
					normals[t+4] = y
					normals[t+5] = z
					normals[t+6] = x
					normals[t+7] = y
					normals[t+8] = z
				}
				// -Cap

				// +Front right
				i += 12
				tmpVerts[i+5] = heights[1]
				tmpVerts[i+9] = heights[0]
				i += 12
				tmpVerts[i+9] = heights[1]
				// -Front right

				// +Front left
				i += 12
				tmpVerts[i+5] = heights[0]
				tmpVerts[i+9] = heights[2]
				i += 12
				tmpVerts[i+5] = heights[2]
				// -Front left

				// +Back
				i += 12
				tmpVerts[i+1] = heights[1]
				i += 12
				tmpVerts[i+1] = heights[1]
				tmpVerts[i+9] = heights[2]
				// -Back
			}
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(tmpVerts)*4, gl.Ptr(tmpVerts), gl.DYNAMIC_DRAW)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
		gl.BufferData(gl.ARRAY_BUFFER, len(normals)*4, gl.Ptr(normals), gl.DYNAMIC_DRAW)
		// -Update

		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(tmpVerts)))
		// -Draw geom

		window.SwapBuffers()
		glfw.PollEvents()

		frames++
		if time-tick >= 1.0 {
			fmt.Printf("FPS: %d\n", frames)
			frames = 0
			tick = time
		}
	}
}
