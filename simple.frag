#version 330 core

in vec3 col;
out vec4 fragment;

void main() {
	fragment = vec4(col.rgb, 1.0);
}
