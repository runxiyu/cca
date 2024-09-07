/* SPDX-License-Identifier: CC0-1.0 */

function humanize_size(bytes) {
	if (Math.abs(bytes) < 1024) {
		return bytes + ' B';
	}
	const units = [' KiB', ' MiB', ' GiB'];
	let u = -1;
	const r = 10;
	do {
		bytes /= 1024;
		++u;
	} while (Math.round(Math.abs(bytes) * r) / r >= 1024 &&
		u < units.length - 1);
	return bytes.toFixed(1) + units[u];
}

const uploadField = document.getElementById("fileupload");
uploadField.onchange = function() {
	if (this.files[0].size > max_file_size) {
		alert(`File size ${humanize_size(this.files[0].size)} ` +
			`exceeds ${humanize_size(max_file_size)}. ` +
			`Either submit a smaller file, ` +
			`or use a file hosting service and submit the URL.`);
		this.value = "";
	}
};
