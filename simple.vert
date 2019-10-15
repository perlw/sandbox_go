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
	/*fcolor.r = 0.25 * (vert.y / 4.0);
	fcolor.g = 0.75 * (vert.y / 4.0);
	fcolor.b = 0.5 * (vert.y / 4.0);*/
	if (normal.y > 0.5) {
		vec3 n = normalize(normal);
		vec3 light = normalize(vec3(-4.0, 4.0, -10.0) - vert);
		float lambertian = max(dot(light, normal), 0.0);
		float specular = 0.0;

		vec3 col = vec3(0.0, 0.0, 0.0);
		if (lambertian > 0.0) {
			vec3 viewDir = normalize(-vert);
			vec3 halfDir = normalize(light + viewDir);
			float specAngle = max(dot(halfDir, normal), 0.0);
			specular = pow(specAngle, 32.0);

			col = (vec3(1.0, 1.0, 1.0) * lambertian) + (specular * vec3(0.8, 0.7, 0.2));
			col *= 2;
		}

		fcolor = (color + col) / 2;
	} else {
		fcolor = (color / 4) * (vert.y / 2.0);
	}

	gl_Position = pMatrix * mvMatrix * vec4(vert.xyz, 1.0);
}
