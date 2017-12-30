#version 330 core

in vec4 vertex;
in vec3 normal;

uniform mat4 mvMatrix;
uniform mat4 pMatrix;

uniform float time;

out vec3 eye;
//out vec3 color;
out vec3 fnormal;
void main() {
	vec3 vert = vertex.xyz;
	if (vert.y > 0.0) {
		vert.y += sin(((vert.x + vert.z) + time) * 0.25);
	}

	/*color.r = vert.x / 18.0;
	color.g = vert.y / 4.0;
	color.b = vert.z / 26.0;
	if (vertex.w > 0.0) {
		color.rgb *= 1.2;
	}*/
	fnormal = normal;

	eye = vec3(0.0, 0.0, 0.0) - vert.xyz;

	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
