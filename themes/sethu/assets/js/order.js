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

			const options = {
				key: jsonResp.VRzpKeyID,
				amount: jsonResp.IAmount,
				currency: 'INR',
				name: 'Sethu Child Development and Family Guidance', //your business name
				description: `Donation towards ${jsonResp.VProject}`,
				image: '',
				order_id: jsonResp.VRzpOrderID,
				callback_url: 'https://sethu.in/sethupay/paid',
				handler: function (response) {
					console.log(response);
				},
				prefill: {
					name: jsonResp.VName, //your customer's name
					email: jsonResp.VEmail,
				},
				notes: {
					project: jsonResp.VProject,
				},
				theme: {
					color: '#3399cc',
				},
			};
			console.log(options);

			var rzp = new Razorpay(options);
			rzp.on('payment.failed', function (response) {
				alert(response.error.code);
				alert(response.error.description);
				alert(response.error.source);
				alert(response.error.step);
				alert(response.error.reason);
				alert(response.error.metadata.order_id);
				alert(response.error.metadata.payment_id);
			});
			rzp.open();
		} catch (e) {
			console.error(e);
		}
	}
})();
