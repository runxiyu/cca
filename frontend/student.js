/*
 * Copyright (c) 2024 Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

document.addEventListener("DOMContentLoaded", () => {
	const socket = new WebSocket("wss://localhost.runxiyu.org:8080/ws")

	/*
	 * TODO I want to make this easily configurable somehow, but I'm unsure
	 * how to fill things into JavaScript. A few possible solutions:
	 * - Replace this string during build time
	 *   This is suboptimal because users should be able to replace it
	 *   during runtime, as the binary is supposed to be decoupled from
	 *   particular instances.
	 * - Replace this string while setting the static handler.
	 *   This is a bit more involved because it requires messing with io.fs;
	 *   I also don't know a way to cleanly escape it.
	 * - Indicate this string somewhere in the template (perhaps via
	 *   a JavaScript variable that we could access).
	 *   This is probably the way to go, especially since html/template
	 *   provides contextual escaping.
	 */

	socket.addEventListener("open", function() {
		let gstate = 0
		let _handleMessage = event => {
			let msg = new String(event?.data)

			/*
			 * Standard IRC Message format parsing without IRCv3
			 * tags or prefixes.  It's a simple enough protocol
			 * format suitable for our use-case.  No need for
			 * protobuf or anything else nontrivial.
			 */
			let mar = msg.split(" ")
			for (let i = 0; i < mar.length; i++) {
				if (mar[i].startsWith(":")) {
					if (i === mar.length - 1) {
						mar[i] = mar[i].substring(1)
						break
					}
					mar[i] = mar[i].substring(1) + " " +
						mar.slice(i + 1).join(" ")
					mar.splice(i + 1)
					break
				}
			}

			switch (mar[0]) {
			case "E": /* unexpected error */
				alert(`The server reported an unexpected error, "${ mar[1] }". The system might be in an inconsistent state.`)
				break
			case "HI":
				document.querySelectorAll(".need-connection").
					forEach(c => {
						c.style.display = "block"
					})
				document.querySelectorAll(".before-connection").
					forEach(c => {
						c.style.display = "none"
					})
				if (mar[1] !== "") {
					let courseIDs = mar[1].split(",")
					for (let i = 0; i < courseIDs.length; i++) {
						document.getElementById(
							`tick${ courseIDs[i] }`
						).checked = true
						if (gstate === 1) {
							document.getElementById(
								`tick${ courseIDs[i] }`
							).disabled = false
						}
					}
				}
				break
			case "U": /* unauthenticated */
				/* TODO: replace this with a box on screen */
				alert("Your session is broken or has expired. You are unauthenticated and the server will reject your commands.")
				break
			case "N":
				document.getElementById(`tick${ mar[1] }`).
					checked = false
				document.getElementById(`tick${ mar[1] }`).
					indeterminate = false
				break
			case "M":
				document.getElementById(`selected${ mar[1] }`).
					textContent = mar[2]
				if (
					mar[2] === document.getElementById(`max${ mar[1] }`).textContent &&
					!(document.getElementById(`tick${ mar[1] }`).checked)
				) {
					document.getElementById(`tick${ mar[1] }`).disabled = true
				} else if (gstate === 1) {
					document.getElementById(`tick${ mar[1] }`).disabled = false
				}
				break
			case "R": /* course selection rejected */
				document.getElementById(`coursestatus${ mar[1] }`).
					textContent = mar[2]
				document.getElementById(`coursestatus${ mar[1] }`).
					style.color = "red"
				document.getElementById(`tick${ mar[1] }`).
					checked = false
				document.getElementById(`tick${ mar[1] }`).
					indeterminate = false
				if (mar[2] === "Full") {
					document.getElementById(`tick${ mar[1] }`).
						disabled = true
				}
				break
			case "Y": /* course selection approved */
				document.getElementById(`coursestatus${ mar[1] }`).
					textContent = ""
				document.getElementById(`coursestatus${ mar[1] }`).
					style.removeProperty("color")
				document.getElementById(`tick${ mar[1] }`).
					checked = true
				document.getElementById(`tick${ mar[1] }`).
					indeterminate = false
				break
			case "STOP":
				gstate = 0
				document.getElementById("stateindicator").textContent = "disabled"
				document.getElementById("confirmbutton").disabled = true
				document.querySelectorAll(".coursecheckbox").forEach(c => {
					c.disabled = true
				})
				break
			case "START":
				gstate = 1
				document.querySelectorAll(".courseitem").forEach(c => {
					if (c.querySelector(".selected-number").textContent !== c.querySelector(".max-number").textContent || c.querySelector(".coursecheckbox").checked) {
						c.querySelector(".coursecheckbox").disabled = false
					}
				})
				document.getElementById("confirmbutton").disabled = false
				document.getElementById("stateindicator").textContent = "enabled"
				break
			default:
				alert(`Invalid command ${ mar[0] } received from socket. Something is wrong.`)
			}
		}
		socket.addEventListener("message", _handleMessage)
		let _handleClose = _event => {
			document.querySelectorAll(".need-connection").forEach(c => {
				c.style.display = "none"
			})
			document.querySelectorAll(".broken-connection").
				forEach(c => {
					c.style.display = "block"
				})
		}
		socket.addEventListener("close", _handleClose)
		socket.send("HELLO")
	})

	document.querySelectorAll(".coursecheckbox").forEach(c => {
		c.addEventListener("input", () => {
			if (c.id.slice(0, 4) !== "tick") {
				alert(`${ c.id } is not in the correct format.`)
				return false
			}
			switch (c.checked) {
			case true:
				c.indeterminate = true
				document.querySelectorAll(".coursecheckbox").forEach(d => {
					if (d.checked === true && d.dataset.group === c.dataset.group && c.id !== d.id) {
						d.indeterminate = true
						socket.send(`N ${ d.id.slice(4) }`)
					}
				})
				socket.send(`Y ${ c.id.slice(4) }`)
				break
			case false:
				c.indeterminate = true
				socket.send(`N ${ c.id.slice(4) }`)
				break
			default:
				alert(`${ c.id }'s "checked" attribute is ${ c.checked } which is invalid.`)
			}
			return false
		})
	})

	document.getElementById("confirmbutton").addEventListener("click", () => {
		socket.send("C")
	})

	document.querySelectorAll(".script-required").forEach(c => {
		c.style.display = "block"
	})
	document.querySelectorAll(".script-unavailable").forEach(c => {
		c.style.display = "none"
	})
})
