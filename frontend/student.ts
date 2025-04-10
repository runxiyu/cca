/*
 * Copyright (c) 2024 Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

var socket: WebSocket;
var global_state: number;
var user_state: number;

const DOM_STATES: Record<string, string> = {
	need_connection: '.need-connection',
	before_connection: '.before-connection',
	broken_connection: '.broken-connection',
	confirmed: '.confirmed',
	unconfirmed: '.unconfirmed',
	neither_confirmed: '.neither-confirmed',
	script_required: '.script-required',
	script_unavailable: '.script-unavailable'
};

document.addEventListener('DOMContentLoaded', (): void => {
	const ws_url = create_websocket_url();
	socket = new WebSocket(ws_url);

	global_state = 0;
	user_state = 0;

	setup_initial_state();
	socket.addEventListener('open', () => setup_socket_handlers());

	setup_course_checkboxes();
	setup_confirmation_buttons();
});

function create_websocket_url(): string {
	const protocol = window.location.protocol === 'http:' ? 'ws:' : 'wss:';
	const hostname = window.location.hostname;
	const port = window.location.port ? `:${window.location.port}` : '';
	return `${protocol}//${hostname}${port}/ws`;
}

function setup_initial_state(): void {
	toggle_elements(DOM_STATES.script_required, true);
	toggle_elements(DOM_STATES.script_unavailable, false);
}

function toggle_elements(selector: string, show: boolean): void {
	document.querySelectorAll(selector).forEach(element => {
		(element as HTMLElement).style.display = show ? 'block' : 'none';
	});
}

function parse_irc_message(message: string): string[] {
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

function check_requirements_met(): boolean {
	const sport_chosen = parseInt(document.getElementById('Sport-chosen')!.textContent!);
	const sport_required = parseInt(document.getElementById('Sport-required')!.textContent!);
	const non_sport_chosen = parseInt(document.getElementById('Non-sport-chosen')!.textContent!);
	const non_sport_required = parseInt(document.getElementById('Non-sport-required')!.textContent!);

	return sport_chosen >= sport_required && non_sport_chosen >= non_sport_required;
}

function update_confirm_button_state(): void {
	const confirm_button = document.getElementById('confirmbutton') as HTMLButtonElement;
	if (global_state === 1 && check_requirements_met()) {
		confirm_button.disabled = false;
	} else {
		confirm_button.disabled = true;
	}
}

function update_course_counters(course_id: string, increment = true): void {
	const course_type = document.getElementById(`type${course_id}`)!.textContent!;
	const counter_element = document.getElementById(`${course_type}-chosen`)!;
	const current_value = parseInt(counter_element.textContent!);
	counter_element.textContent = String(current_value + (increment ? 1 : -1));

	update_confirm_button_state();
}

function update_confirmed_course_details(handle: string): void {
	const elements = ['name', 'type', 'teacher', 'location'].reduce<Record<string, HTMLElement>>((acc, field) => {
		acc[field] = document.getElementById(`confirmed-${field}-${handle}`)!;
		return acc;
	}, {});

	document.querySelectorAll('.coursecheckbox').forEach(chk => {
		const checkbox = chk as HTMLInputElement;
		if (checkbox.dataset.group === handle && checkbox.checked) {
			elements.name.textContent = checkbox.dataset.title!;
			elements.type.textContent = checkbox.dataset.type!;
			elements.teacher.textContent = checkbox.dataset.teacher!;
			elements.location.textContent = checkbox.dataset.location!;
		}
	});
}

function swap_2_and_3(input: string): string {
	return input.replace(/[23]/g, (char) => {
		if (char === "2") return "3";
		if (char === "3") return "2";
		return char;
	});
}

function handle_course_selection(checkbox: HTMLInputElement): void {
	if (!checkbox.id.startsWith('tick')) {
		alert(`${checkbox.id} is not in the correct format.`);
		return;
	}
	const course_id = checkbox.id.slice(4);
	checkbox.indeterminate = true;

	if (checkbox.checked) {
		let swapped_group = swap_2_and_3(checkbox.dataset.group!);
		document.querySelectorAll('.coursecheckbox').forEach(chk => {
			const other_checkbox = chk as HTMLInputElement;
			if (
				other_checkbox.checked &&
				(other_checkbox.dataset.group === checkbox.dataset.group || other_checkbox.dataset.group === swapped_group) &&
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

function setup_socket_handlers(): void {
	socket.addEventListener('message', event => handle_socket_message(event));
	socket.addEventListener('close', () => {
		toggle_elements(DOM_STATES.need_connection, false);
		toggle_elements(DOM_STATES.broken_connection, true);
	});
	socket.send('HELLO');
}

function handle_hi_message(course_list = ''): void {
	console.log(course_list);

	if (course_list) {
		course_list.split(',').forEach(course_id => {
			const checkbox = document.getElementById(`tick${course_id}`) as HTMLInputElement;
			checkbox.checked = true;
			update_course_counters(course_id, true);
		});
	}

	toggle_elements(DOM_STATES.need_connection, true);
	toggle_elements(DOM_STATES.before_connection, false);

	if (user_state === 1) {
		_handle_confirmation_state();
	}
}

function handle_course_removal(course_id: string): void {
	const checkbox = document.getElementById(`tick${course_id}`) as HTMLInputElement;
	checkbox.checked = false;
	checkbox.indeterminate = false;
	update_course_counters(course_id, false);
}

function handle_course_max_update(course_id: string, selected_count: string): void {
	const selected_element = document.getElementById(`selected${course_id}`)!;
	const max_element = document.getElementById(`max${course_id}`)!;
	const checkbox = document.getElementById(`tick${course_id}`) as HTMLInputElement;

	selected_element.textContent = selected_count;
	checkbox.disabled = selected_count === max_element.textContent && !checkbox.checked;
}

function handle_course_rejection(course_id: string, reason: string): void {
	const status_element = document.getElementById(`coursestatus${course_id}`)!;
	const checkbox = document.getElementById(`tick${course_id}`) as HTMLInputElement;

	status_element.textContent = reason;
	(status_element as HTMLElement).style.color = 'red';
	checkbox.checked = false;
	checkbox.indeterminate = false;
	if (reason === 'Full') {
		checkbox.disabled = true;
	}
	update_confirm_button_state();
}

function handle_course_approval(course_id: string): void {
	const status_element = document.getElementById(`coursestatus${course_id}`)!;
	const checkbox = document.getElementById(`tick${course_id}`) as HTMLInputElement;

	status_element.textContent = '';
	(status_element as HTMLElement).style.removeProperty('color');
	checkbox.checked = true;
	checkbox.indeterminate = false;
	update_course_counters(course_id, true);
}

function handle_stop_state(): void {
	global_state = 0;
	document.getElementById('stateindicator')!.textContent = 'disabled';
	(document.getElementById('confirmbutton') as HTMLButtonElement).disabled = true;
	(document.getElementById('unconfirmbutton') as HTMLButtonElement).disabled = true;
	document.querySelectorAll('.coursecheckbox').forEach(c => {
		(c as HTMLInputElement).disabled = true;
	});
}

function handle_start_state(): void {
	global_state = 1;
	(document.getElementById('unconfirmbutton') as HTMLButtonElement).disabled = false;
	document.getElementById('stateindicator')!.textContent = 'You may choose courses now!';

	document.querySelectorAll('.courseitem').forEach(course => {
		const checkbox = course.querySelector('.coursecheckbox') as HTMLInputElement;
		const selected = course.querySelector('.selected-number')!;
		const max = course.querySelector('.max-number')!;

		checkbox.disabled = !(
			selected.textContent !== max.textContent ||
			checkbox.checked
		);
	});

	update_confirm_button_state();
}

function handle_confirmation_state(): void {
	user_state = 1;
}

function _handle_confirmation_state(): void {
	document.querySelectorAll('.confirmed-handle').forEach(handle => {
		update_confirmed_course_details(handle.textContent!);
	});
	toggle_elements(DOM_STATES.unconfirmed, false);
	toggle_elements(DOM_STATES.confirmed, true);
	toggle_elements(DOM_STATES.neither_confirmed, false);
}

function handle_unconfirmation_state(): void {
	user_state = 0;
	toggle_elements(DOM_STATES.unconfirmed, true);
	toggle_elements(DOM_STATES.confirmed, false);
	toggle_elements(DOM_STATES.neither_confirmed, false);
}

function setup_course_checkboxes(): void {
	document.querySelectorAll('.coursecheckbox').forEach(c => {
		const checkbox = c as HTMLInputElement;
		checkbox.addEventListener('input', () => handle_course_selection(checkbox));
	});
}

function setup_confirmation_buttons(): void {
	(document.getElementById('confirmbutton') as HTMLButtonElement).addEventListener('click', () => socket.send('YC'));
	(document.getElementById('unconfirmbutton') as HTMLButtonElement).addEventListener('click', () => socket.send('NC'));
}

function handle_socket_message(event: MessageEvent): void {
	const message = parse_irc_message(String(event?.data));
	const [command, ...args] = message;

	const message_handlers: Record<string, () => void> = {
		'E': () => alert(args[0]),
		'HI': () => handle_hi_message(...args),
		'U': () => alert('Your session is broken or has expired. You are unauthenticated and the server will reject your commands.'),
		'N': () => handle_course_removal(args[0]),
		'M': () => handle_course_max_update(args[0], args[1]),
		'R': () => handle_course_rejection(args[0], args[1]),
		'Y': () => handle_course_approval(args[0]),
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
