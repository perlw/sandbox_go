#version 330 core

layout(location = 1) in vec3 vertex;
layout(location = 2) in vec3 normal;
layout(location = 3) in vec3 color;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;
uniform mat4 normalMatrix;

out vec3 fcolor;
void main() {
	vec3 vert = vertex.xyz;

	/*
	fcolor.r = vert.x / 18.0;
	fcolor.g = vert.y / 4.0;
	fcolor.b = vert.z / 26.0;
	*/
	fcolor.r = 0.25 * (vert.y / 4.0);
	fcolor.g = 0.75 * (vert.y / 4.0);
	fcolor.b = 0.5 * (vert.y / 4.0);
	if (normal.y > 0.5) {
		fcolor = color;
	}

	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
