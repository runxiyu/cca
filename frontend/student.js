/*
 * Copyright (c) 2024 Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

const DOM_STATES = {
	need_connection: '.need-connection',
	before_connection: '.before-connection',
	broken_connection: '.broken-connection',
	confirmed: '.confirmed',
	unconfirmed: '.unconfirmed',
	neither_confirmed: '.neither-confirmed',
	script_required: '.script-required',
	script_unavailable: '.script-unavailable'
};

document.addEventListener('DOMContentLoaded', () => {
	const ws_url = create_websocket_url();
	window.socket = new WebSocket(ws_url);

	window.global_state = 0;
	window.user_state = 0;

	setup_initial_state();
	socket.addEventListener('open', () => setup_socket_handlers(socket));

	setup_course_checkboxes(socket);
	setup_confirmation_buttons(socket);
});

function create_websocket_url() {
	const protocol = window.location.protocol === 'http:' ? 'ws:' : 'wss:';
	const hostname = window.location.hostname;
	const port = window.location.port ? `:${window.location.port}` : '';
	return `${protocol}//${hostname}${port}/ws`;
}

function setup_initial_state() {
	toggle_elements(DOM_STATES.script_required, true);
	toggle_elements(DOM_STATES.script_unavailable, false);
}

function toggle_elements(selector, show) {
	document.querySelectorAll(selector).forEach(element => {
		element.style.display = show ? 'block' : 'none';
	});
}

function parse_irc_message(message) {
	const parts = message.split(' ');
	for (let i = 0; i < parts.length; i++) {
		if (parts[i].startsWith(':')) {
			if (i === parts.length - 1) {
				parts[i] = parts[i].substring(1);
				break;
			}
			parts[i] = parts[i].substring(1) + ' ' + parts.slice(i + 1).join(' ');
			parts.splice(i + 1);
			break;
		}
	}
	return parts;
}

function check_requirements_met() {
	const sport_chosen = parseInt(document.getElementById('Sport-chosen').textContent);
	const sport_required = parseInt(document.getElementById('Sport-required').textContent);
	const non_sport_chosen = parseInt(document.getElementById('Non-sport-chosen').textContent);
	const non_sport_required = parseInt(document.getElementById('Non-sport-required').textContent);

	return sport_chosen >= sport_required && non_sport_chosen >= non_sport_required;
}

function update_confirm_button_state() {
	const confirm_button = document.getElementById('confirmbutton');
	if (window.global_state === 1 && check_requirements_met()) {
		confirm_button.disabled = false;
	} else {
		confirm_button.disabled = true;
	}
}

function update_course_counters(course_id, increment = true) {
	const course_type = document.getElementById(`type${course_id}`).textContent;
	const counter_element = document.getElementById(`${course_type}-chosen`);
	const current_value = parseInt(counter_element.textContent);
	counter_element.textContent = current_value + (increment ? 1 : -1);

	update_confirm_button_state();
}

function update_confirmed_course_details(handle) {
	const elements = ['name', 'type', 'teacher', 'location'].reduce((acc, field) => {
		acc[field] = document.getElementById(`confirmed-${field}-${handle}`);
		return acc;
	}, {});

	document.querySelectorAll('.coursecheckbox').forEach(checkbox => {
		if (checkbox.dataset.group === handle && checkbox.checked) {
			elements.name.textContent = checkbox.dataset.title;
			elements.type.textContent = checkbox.dataset.type;
			elements.teacher.textContent = checkbox.dataset.teacher;
			elements.location.textContent = checkbox.dataset.location;
		}
	});
}

function handle_course_selection(socket, checkbox) {
	if (!checkbox.id.startsWith('tick')) {
		alert(`${checkbox.id} is not in the correct format.`);
		return;
	}
	const course_id = checkbox.id.slice(4);
	checkbox.indeterminate = true;

	if (checkbox.checked) {
		document.querySelectorAll('.coursecheckbox').forEach(other_checkbox => {
			if (
				other_checkbox.checked &&
				other_checkbox.dataset.group === checkbox.dataset.group &&
				other_checkbox.id !== checkbox.id
			) {
				other_checkbox.indeterminate = true;
				socket.send(`N ${other_checkbox.id.slice(4)}`);
			}
		});
		socket.send(`Y ${course_id}`);
	} else {
		socket.send(`N ${course_id}`);
	}
}

function setup_socket_handlers(socket) {
	socket.addEventListener('message', event => handle_socket_message(socket, event));
	socket.addEventListener('close', () => {
		toggle_elements(DOM_STATES.need_connection, false);
		toggle_elements(DOM_STATES.broken_connection, true);
	});
	socket.send('HELLO');
}

function handle_socket_message(socket, event) {
	const message = parse_irc_message(String(event?.data));
	const [command, ...args] = message;

	const message_handlers = {
		'E': () => alert(args[0]),
		'HI': () => handle_hi_message(...args),
		'U': () => alert('Your session is broken or has expired. You are unauthenticated and the server will reject your commands.'),
		'N': () => handle_course_removal(args[0]),
		'M': () => handle_course_max_update(...args),
		'R': () => handle_course_rejection(...args),
		'Y': () => handle_course_approval(...args),
		'STOP': () => handle_stop_state(),
		'START': () => handle_start_state(),
		'YC': () => handle_confirmation_state(),
		'NC': () => handle_unconfirmation_state(),
		'RC': () => alert(args[0])
	};

	const handler = message_handlers[command];
	if (handler) {
		handler();
	} else {
		alert(`Invalid command ${command} received from socket. Something is wrong.`);
	}
}

function handle_hi_message(course_list = '') {
	toggle_elements(DOM_STATES.need_connection, true);
	toggle_elements(DOM_STATES.before_connection, false);

	if (course_list) {
		course_list.split(',').forEach(course_id => {
			const checkbox = document.getElementById(`tick${course_id}`);
			checkbox.checked = true;
			update_course_counters(course_id, true);
		});
	}
}

function handle_course_removal(course_id) {
	const checkbox = document.getElementById(`tick${course_id}`);
	checkbox.checked = false;
	checkbox.indeterminate = false;
	update_course_counters(course_id, false);
}

function handle_course_max_update(course_id, selected_count) {
	const selected_element = document.getElementById(`selected${course_id}`);
	const max_element = document.getElementById(`max${course_id}`);
	const checkbox = document.getElementById(`tick${course_id}`);

	selected_element.textContent = selected_count;
	checkbox.disabled = selected_count === max_element.textContent && !checkbox.checked;
}

function handle_course_rejection(course_id, reason) {
	const status_element = document.getElementById(`coursestatus${course_id}`);
	const checkbox = document.getElementById(`tick${course_id}`);

	status_element.textContent = reason;
	status_element.style.color = 'red';
	checkbox.checked = false;
	checkbox.indeterminate = false;
	if (reason === 'Full') {
		checkbox.disabled = true;
	}
	update_confirm_button_state();
}

function handle_course_approval(course_id) {
	const status_element = document.getElementById(`coursestatus${course_id}`);
	const checkbox = document.getElementById(`tick${course_id}`);

	status_element.textContent = '';
	status_element.style.removeProperty('color');
	checkbox.checked = true;
	checkbox.indeterminate = false;
	update_course_counters(course_id, true);
}

function handle_stop_state() {
	window.global_state = 0;
	document.getElementById('stateindicator').textContent = 'disabled';
	document.getElementById('confirmbutton').disabled = true;
	document.getElementById('unconfirmbutton').disabled = true;
	document.querySelectorAll('.coursecheckbox').forEach(c => {
		c.disabled = true;
	});
}

function handle_start_state() {
	window.global_state = 1;
	document.getElementById('unconfirmbutton').disabled = false;
	document.getElementById('stateindicator').textContent = 'enabled';

	document.querySelectorAll('.courseitem').forEach(course => {
		const checkbox = course.querySelector('.coursecheckbox');
		const selected = course.querySelector('.selected-number');
		const max = course.querySelector('.max-number');

		checkbox.disabled = !(
			selected.textContent !== max.textContent ||
			checkbox.checked
		);
	});

	update_confirm_button_state();
}

function handle_confirmation_state() {
	window.user_state = 1;
	document.querySelectorAll('.confirmed-handle').forEach(handle => {
		update_confirmed_course_details(handle.textContent);
	});
	toggle_elements(DOM_STATES.unconfirmed, false);
	toggle_elements(DOM_STATES.confirmed, true);
	toggle_elements(DOM_STATES.neither_confirmed, false);
}

function handle_unconfirmation_state() {
	window.user_state = 0;
	toggle_elements(DOM_STATES.unconfirmed, true);
	toggle_elements(DOM_STATES.confirmed, false);
	toggle_elements(DOM_STATES.neither_confirmed, false);
}

function setup_course_checkboxes(socket) {
	document.querySelectorAll('.coursecheckbox').forEach(checkbox => {
		checkbox.addEventListener('input', () => handle_course_selection(socket, checkbox));
	});
}

function setup_confirmation_buttons(socket) {
	document.getElementById('confirmbutton').addEventListener('click', () => socket.send('YC'));
	document.getElementById('unconfirmbutton').addEventListener('click', () => socket.send('NC'));
}
