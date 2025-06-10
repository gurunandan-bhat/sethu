(function () {
	const form = document.querySelector('#donate');
	form.addEventListener('submit', (event) => {
		event.preventDefault();
		console.log(event);
		getOrderID();
	});

	async function getOrderID() {
		const project = event.submitter.dataset['project'];
		const orderURL = event.target.dataset['orderUrl'];

		const formData = new FormData(event.target);
		formData.append('project', project);

		try {
			const response = await fetch(orderURL, {
				method: 'POST',
				body: formData,
			});
			if (!response.ok) {
				throw new Error(`Resonse.status: ${response.status}`);
			}
			const jsonResp = await response.json();
			console.log(jsonResp);
		} catch (e) {
			console.error(e);
		}
	}
})();
