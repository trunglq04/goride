package emailtemplates

var OtpSendingTemplate string = `
		<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px;">
			<div style="text-align: center; margin-bottom: 30px;">
				<h1 style="color: #1a1a2e; font-size: 28px; margin: 0;">GoRide</h1>
				<p style="color: #666; margin-top: 8px;">Your ride, your way</p>
			</div>
			<div style="background: #f8f9fa; border-radius: 12px; padding: 30px; text-align: center;">
				<h2 style="color: #333; font-size: 20px; margin-bottom: 10px;">Verify your email address</h2>
				<p style="color: #666; margin-bottom: 20px;">Enter the following code to complete your registration:</p>
				<div style="background: #1a1a2e; color: #fff; font-size: 36px; font-weight: bold; letter-spacing: 8px; padding: 20px 30px; border-radius: 8px; display: inline-block;">
					%s
				</div>
				<p style="color: #999; font-size: 13px; margin-top: 20px;">This code expires in 10 minutes.</p>
			</div>
			<p style="color: #999; font-size: 12px; text-align: center; margin-top: 30px;">
				If you didn't create a GoRide account, you can safely ignore this email.
			</p>
		</div>
	`
