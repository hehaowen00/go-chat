<script>
	import { formatDate } from '$lib';
	import { writable } from 'svelte/store';

	// @ts-ignore
	let channelName = '';
	let currentChannel = '';

	/**
	 * @type {EventSource | undefined}
	 */
	let eventSource = undefined;
	let history = writable([]);
	let message = '';

	async function getChannels() {
		let res = await fetch('http://localhost:8080/channels');
		if (!res.ok) {
			return [];
		}

		return await res.json();
	}

	async function newChannel() {
		let res = await fetch('http://localhost:8080/channels/new', {
			method: 'POST',
			body: JSON.stringify({ name: channelName })
		});

		if (!res.ok) {
			console.log(res);
		}

		return await res.json();
	}

	/**
	 * @param {string} e
	 */
	function changeChannel(e) {
		if (!e) {
			if (eventSource) {
				eventSource.close();
				eventSource = undefined;
			}

			history.set([]);
			return;
		}

		if (eventSource) {
			eventSource.close();
		}

		currentChannel = e;

		console.log(currentChannel);

		eventSource = undefined;
		eventSource = new EventSource('http://localhost:8080/chat/' + currentChannel);

		eventSource.onmessage = (event) => {
			let msg = JSON.parse(event.data);
			// @ts-ignore
			history.update((history) => [...history, msg]);
		};

		eventSource.onopen = () => {
			history.set([]);
		};
	}

	async function sendMessage() {
		let res = await fetch('http://localhost:8080/send/' + currentChannel, {
			method: 'POST',
			// @ts-ignore
			body: new FormData(document.querySelector('form'))
		});
		if (!res.ok) {
			console.log(res);
		}

		message = '';
	}

	let value = '';
	// @ts-ignore
	$: changeChannel(value);
</script>

<div>
	<h1>Chat</h1>

	<div>
		<input type="text" bind:value={channelName} />
		<button on:click={newChannel}>Add Channel</button>
	</div>

	<select bind:value on:selectionchange={changeChannel}>
		<option value="">Select Channel...</option>
		{#await getChannels() then channels}
			{#each channels as channel}
				<option value={channel.id}>{channel.name}</option>
			{/each}
		{/await}
	</select>

	<div class="messages">
		{#each $history as message}
			<p>
				[{formatDate(message.timestamp)}]
				{#if message.user}
					{message.user}:
				{/if}
				{message.content}
				{#if message.image}
					<br />
					<!-- svelte-ignore a11y-missing-attribute -->
					<img src="http://localhost:8080{message.image}" />
				{/if}
				{#if message.file}
					<a href="http://localhost:8080{message.file.url}">{message.file.filename}</a>
				{/if}
			</p>
		{/each}
	</div>

	<form class="send-message-form" enctype="multipart/form-data">
		<textarea name="message" bind:value={message} rows="5"></textarea>
		<input type="file" name="upload" />
		<button type="button" on:click={sendMessage}>Send</button>
	</form>
</div>

<style>
	h1,
	p {
		margin-block-start: 0.15rem;
		margin-block-end: 0.15rem;
	}

	img {
		max-width: 25%;
	}

	.messages {
		margin-top: 5px;
	}

	.send-message-form {
		display: flex;
		flex-direction: column;
	}

	.send-message-form > * + * {
		margin-top: 5px;
	}
</style>
