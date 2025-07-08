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

		var response;
		try {
			response = await fetch(orderURL, {
				method: 'POST',
				body: formData,
			});
		} catch (err) {
			console.log('Error fetching order id: ', err);
			return;
		}
		if (!response.ok) {
			const err = await response.json();
			console.log(err);
			return;
		}

		const rzpResp = await response.json();
		const options = {
			key: rzpResp.VRzpKeyID,
			amount: rzpResp.IAmount,
			currency: 'INR',
			name: 'Sethu Child Development and Family Guidance', //your business name
			description: `Donation towards ${rzpResp.VProject}`,
			image: '',
			order_id: rzpResp.VRzpOrderID,
			callback_url: callbackURL,
			redirect: true,
			remember_customer: false,
			handler: function (response) {
				console.log(response);
			},
			prefill: {
				name: rzpResp.VName,
				email: rzpResp.VEmail,
			},
			theme: {
				color: '#3399cc',
			},
		};
		var rzp = new Razorpay(options);

		try {
			rzp.open();
		} catch (err) {
			console.log('Error starting payment modal', err);
			return;
		}
	}
})();
