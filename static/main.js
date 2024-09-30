/*
 * Copyright (c) 2024 Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

var connect = function(socket) {
	var _handleMessage = event => {
		let msg = new String(event?.data)

		/*
		 * Standard IRC Message format parsing without IRCv3 tags or prefixes.
		 * It's a simple enough protocol format suitable for our use-case.
		 * No need for protobuf or anything else nontrivial.
		 */
		let mar = msg.split(" ")
		for (let i = 0; i < mar.length; i++) {
			if (mar[i].startsWith(":")) {
				if (i === mar.length - 1) {
					mar[i] = mar[i].substring(1)
					break
				}
				mar[i] = mar[i].substring(1) + " " + mar.slice(i + 1).join(" ")
				mar.splice(i + 1)
				break
			}
		}

		switch (mar[0]) {
		case "E": /* unexpected error */
			alert(`The server reported an unexpected error, "${ mar[1] }". The system might be in an inconsistent state.`)
			break
		case "HI":
			document.querySelectorAll(".need-connection").forEach(c => {
				c.style.display = "block"
			})
			document.querySelectorAll(".before-connection").forEach(c => {
				c.style.display = "none"
			})
			break
		case "U": /* unauthenticated */
			/* TODO: replace this with a box on screen */
			alert("Your session is broken or has expired. You are unauthenticated and the server will reject your commands.")
			break
		case "N":
			document.getElementById(`selected${ mar[1] }`).textContent = mar[2]
			break
		case "R": /* course selection rejected */
			document.getElementById(`coursestatus${ mar[1] }`).textContent = mar[2]
			document.getElementById(`coursestatus${ mar[1] }`).style.color = "red"
			document.getElementById(`tick${ mar[1] }`).checked = false
			document.getElementById(`tick${ mar[1] }`).indeterminate = false
			break
		case "Y": /* course selection approved */
			document.getElementById(`coursestatus${ mar[1] }`).textContent = ""
			document.getElementById(`coursestatus${ mar[1] }`).style.removeProperty("color")
			document.getElementById(`tick${ mar[1] }`).checked = true
			document.getElementById(`tick${ mar[1] }`).indeterminate = false
			break
		default:
			alert(`Invalid command ${ mar[0] } received from socket. Something is wrong.`)
		}
	}
	socket.addEventListener("message", _handleMessage)
	var _handleClose = _event => {
		document.querySelectorAll(".need-connection").forEach(c => {
			c.style.display = "none"
		})
		document.querySelectorAll(".broken-connection").forEach(c => {
			c.style.display = "block"
		})
	}
	socket.addEventListener("close", _handleClose)
	socket.send("HELLO")
}

const socket = new WebSocket("ws://localhost:5555/ws")
socket.addEventListener("open", function() {
	connect(socket)
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
			socket.send(`Y ${ c.id.slice(4) }`)
			break
		case false:
			c.indeterminate = false
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
