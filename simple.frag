#version 330 core

in vec4 feye;
in vec3 fvert;
in vec3 fcolor;
in vec3 fnormal;
out vec4 fragment;

void main() {
	vec3 normal = normalize(fnormal);
	vec3 light = normalize(vec3(-4.0, 4.0, -10.0) - fvert);
	float lambertian = max(dot(light, normal), 0.0);
	float specular = 0.0;

	vec3 frag = vec3(0.0, 0.0, 0.0);
	if (lambertian > 0.0) {
		vec3 viewDir = normalize(-fvert);
		vec3 halfDir = normalize(light + viewDir);
		float specAngle = max(dot(halfDir, normal), 0.0);
		specular = pow(specAngle, 32.0);

		frag = (fcolor * lambertian) + (specular * vec3(0.8, 0.7, 0.2));
	}

	fragment = vec4(frag.rgb, 1.0);
}
