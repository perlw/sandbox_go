#version 330 core

in float depth;
in vec3 col;
out vec4 fragment;

void main() {
	vec4 color = vec4(col.rgb, 1.0);
	color.r = col.r * depth;
	color.g = col.g * depth;
	color.b = col.b * depth;
	fragment = color;
}
