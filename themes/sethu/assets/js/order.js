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
		const callbackURL = event.target.dataset['callbackUrl'];

		const formData = new FormData(event.target);
		formData.append('project', project);

		try {
			const response = await fetch(orderURL, {
				method: 'POST',
				body: formData,
			});
			if (!response.ok) {
				data = await response.json();
				throw data;
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
				callback_url: callbackURL,
				redirect: true,
				remember_customer: false,
				handler: function (response) {
					console.log(response);
				},
				prefill: {
					name: jsonResp.VName,
					email: jsonResp.VEmail,
				},
				theme: {
					color: '#3399cc',
				},
			};
			console.log(options);
			var rzp = new Razorpay(options);
			rzp.open();
		} catch (e) {
			// const err = e.json().then((json) => console.log(json));
			console.log(e);
		}
	}
})();
