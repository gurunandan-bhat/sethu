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
				handler: function (response) {
					alert(response.razorpay_payment_id);
					alert(response.razorpay_order_id);
					alert(response.razorpay_signature);
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
			console.log('Response: ', response);

			const reader = response.body.getReader();
			while (true) {
				const { done, value } = await reader.read();
				if (done) {
					break;
				}
				console.log('Chunk: ', value);
			}
		} catch (e) {
			console.error(e);
		}
	}
})();
