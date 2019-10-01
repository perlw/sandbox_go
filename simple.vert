#version 330 core

in vec3 vertex;
in vec3 normal;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;
uniform mat4 normalMatrix;

out vec4 feye;
out vec3 fvert;
out vec3 fcolor;
out vec3 fnormal;
void main() {
	vec3 vert = vertex.xyz;

	fcolor.r = vert.x / 18.0;
	fcolor.g = vert.y / 4.0;
	fcolor.b = vert.z / 26.0;
	/*if (vertex.w > 0.0) {
		fcolor.rgb *= 1.2;
	}*/

	feye = mvMatrix * vec4(0.0, 0.0, 0.0, 1.0);
	vec4 vert4 = mvMatrix * vec4(vert.xyz, 1.0);
	fvert = vec3(vert4) / vert4.w;
	// fcolor = vec3(0.1, 0.1, 0.1);
	fnormal = vec3(normalMatrix * vec4(normal, 0.0)).xyz;

	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
