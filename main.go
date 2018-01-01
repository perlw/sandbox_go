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
	"time"
)

func init() {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(2)
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

func calcNormal(p1, p2, p3 mgl32.Vec3) mgl32.Vec3 {
	u := p2.Sub(p1)
	v := p3.Sub(p1)
	return u.Cross(v)
}

type Mesh struct {
	verts   []mgl32.Vec3
	normals []mgl32.Vec3
}

func generateMesh(vertFunc func(x, y float32, vertex mgl32.Vec3) mgl32.Vec3) Mesh {
	width := 32
	height := 48

	mesh := Mesh{
		verts:   []mgl32.Vec3{},
		normals: []mgl32.Vec3{},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var rootVerts []mgl32.Vec3

			offsetx := float32(0.0)
			offsety := float32(0.0)
			if y%2 == 0 {
				origox := float32(x) * 1.1
				origoy := -(float32(y) / 2.0) * 1.1

				rootVerts = []mgl32.Vec3{
					// Top
					{origox + offsetx + 0.0, 2.0, origoy + offsety + 0.5},
					{origox + offsetx + 0.5, 2.0, origoy + offsety - 0.5},
					{origox + offsetx - 0.5, 2.0, origoy + offsety - 0.5},

					// Bottom
					{origox + 0.0, 0.0, origoy + 0.5},
					{origox + 0.5, 0.0, origoy - 0.5},
					{origox - 0.5, 0.0, origoy - 0.5},
				}
			} else {
				origox := (float32(x) + 0.5) * 1.1
				origoy := -((float32(y) / 2.0) - 0.5) * 1.1

				rootVerts = []mgl32.Vec3{
					// Top
					{origox + offsetx - 0.0, 2.0, origoy + offsety - 0.5},
					{origox + offsetx - 0.5, 2.0, origoy + offsety + 0.5},
					{origox + offsetx + 0.5, 2.0, origoy + offsety + 0.5},

					// Bottom
					{origox - 0.0, 0.0, origoy - 0.5},
					{origox - 0.5, 0.0, origoy + 0.5},
					{origox + 0.5, 0.0, origoy + 0.5},
				}
			}
			if vertFunc != nil {
				for i, vert := range rootVerts {
					rootVerts[i] = vertFunc(float32(x), float32(y), vert)
				}
			}
			rootNormals := []mgl32.Vec3{
				// Cap
				calcNormal(rootVerts[0], rootVerts[1], rootVerts[2]),

				// Front right
				calcNormal(rootVerts[3], rootVerts[1], rootVerts[0]),

				// Front left
				calcNormal(rootVerts[3], rootVerts[0], rootVerts[2]),

				// Back
				calcNormal(rootVerts[1], rootVerts[4], rootVerts[5]),
			}

			mesh.verts = append(mesh.verts, []mgl32.Vec3{
				// +Cap
				rootVerts[0],
				rootVerts[1],
				rootVerts[2],
				// -Cap

				// +Front right
				rootVerts[3],
				rootVerts[1],
				rootVerts[0],

				rootVerts[3],
				rootVerts[4],
				rootVerts[1],
				// -Front right

				// +Front left
				rootVerts[3],
				rootVerts[0],
				rootVerts[2],

				rootVerts[3],
				rootVerts[2],
				rootVerts[5],
				// -Front left

				// +Back
				rootVerts[1],
				rootVerts[4],
				rootVerts[5],

				rootVerts[1],
				rootVerts[5],
				rootVerts[2],
				// -Back
			}...)

			mesh.normals = append(mesh.normals, []mgl32.Vec3{
				// +Cap
				rootNormals[0],
				rootNormals[0],
				rootNormals[0],
				// -Cap

				// +Front right
				rootNormals[1],
				rootNormals[1],
				rootNormals[1],

				rootNormals[1],
				rootNormals[1],
				rootNormals[1],
				// -Front right

				// +Front left
				rootNormals[2],
				rootNormals[2],
				rootNormals[2],

				rootNormals[2],
				rootNormals[2],
				rootNormals[2],
				// -Front left

				// +Back
				rootNormals[3],
				rootNormals[3],
				rootNormals[3],

				rootNormals[3],
				rootNormals[3],
				rootNormals[3],
				// -Back
			}...)
		}
	}
	//fmt.Printf("%v\n", baseVerts)

	return mesh
}

const width = 1280
const height = 720

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

	window, err := glfw.CreateWindow(width, height, "Testing", nil, nil)
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
	gl.Viewport(0, 0, width, height)
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
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/float32(height), 1.0, 100.0)
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

	var vbo uint32
	var vbo2 uint32
	baseMesh := generateMesh(nil)
	{
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, (len(baseMesh.verts)*3)*4, gl.Ptr(baseMesh.verts), gl.DYNAMIC_DRAW)

		vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertex\x00")))
		gl.EnableVertexAttribArray(vertAttrib)
		gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.GenBuffers(1, &vbo2)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
		gl.BufferData(gl.ARRAY_BUFFER, (len(baseMesh.normals)*3)*4, gl.Ptr(baseMesh.normals), gl.DYNAMIC_DRAW)

		normalAttrib := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
		gl.EnableVertexAttribArray(normalAttrib)
		gl.VertexAttribPointer(normalAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	}
	// -Setup geom

	fmt.Printf("Polys: %d | Vertices: %d | Normals: %d\n", len(baseMesh.verts)/3, len(baseMesh.verts), len(baseMesh.normals))

	waveRebuild := make(chan Mesh)
	go func(ch chan Mesh) {
		tickRate := int64(1000000000) / int64(60000000)

		curr := time.Now().UnixNano()
		for {
			tick := time.Now().UnixNano()
			if tick-curr >= tickRate {
				curr = tick

				time := float64(tick) / 1000000000
				mesh := generateMesh(func(x, y float32, vertex mgl32.Vec3) mgl32.Vec3 {
					if vertex[1] > 0.0 {
						// Base height
						vertex[1] += float32((math.Sin(float64(x)+time) - math.Cos(float64(y)+time)) * 0.25)

						// Tweaked height
						vertex[1] += float32(math.Sin(float64(vertex[0]+vertex[2])+time) * 0.25)
					}

					return vertex
				})
				ch <- mesh
			}

			runtime.Gosched()
		}
	}(waveRebuild)

	var tick float64 = 0.0
	var frames uint32 = 0
	for !window.ShouldClose() {
		time := glfw.GetTime()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		select {
		case mesh := <-waveRebuild:
			gl.BindVertexArray(vao)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.BufferData(gl.ARRAY_BUFFER, (len(mesh.verts)*3)*4, gl.Ptr(mesh.verts), gl.DYNAMIC_DRAW)
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo2)
			gl.BufferData(gl.ARRAY_BUFFER, (len(mesh.normals)*3)*4, gl.Ptr(mesh.normals), gl.DYNAMIC_DRAW)
		default:
		}

		// +Draw geom
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(baseMesh.verts)))
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
