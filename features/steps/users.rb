require 'net/http'
require 'test/unit'
require 'minitest/autorun'

@status_code = 0

def sendRequest(url, params)
	uri = URI(url)
	req = Net::HTTP::Post.new(uri, 'Content-Type' => 'application/json')
	req.body = params
	res = Net::HTTP.start(uri.hostname, uri.port) do |http|
			http.request(req)
	end

	return res
end

Given("I do have a running web server") do
end

When("I sign up with {word} credentials") do |isValid|
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			   username = 'Test'
		else
			   username = 'a'
		end

		url = 'http://0.0.0.0:3000/users'
		body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
		res = sendRequest(url, body)
		@status_code = Integer(res.code)
end

Then("I should see a {int} response") do |code|
		assert_equal(code, @status_code)
end
