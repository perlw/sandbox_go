#version 330 core

in vec3 eye;
//in vec3 color;
in vec3 fnormal;
out vec4 fragment;

void main() {
	vec3 normal = normalize(fnormal);
	vec3 light = normalize(vec3(1.0, 1.0, 1.0));
	float ndotl = max(dot(normal, light), 0.0);
	vec3 color = vec3(0.0, 0.0, 0.0);
	vec3 frag = color.rgb;

	if (ndotl > 0.0) {
		vec3 H = normalize(eye + vec3(2.0, 8.0, 0.0));
		float shine = pow(max(0.0, dot(normal, H)), 32.0);
		frag.rgb += (color.rgb * (ndotl + 0.25)) + (shine * vec3(1.0, 0.8, 0.0));
	}
	fragment = vec4(frag.rgb, 1.0);
}
