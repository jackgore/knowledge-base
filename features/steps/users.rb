require 'net/http'
require 'test/unit'
require 'minitest/autorun'

@status_code = 0

def sendPostRequest(url, params)
	uri = URI(url)
	req = Net::HTTP::Post.new(uri, 'Content-Type' => 'application/json')
	req.body = params
	res = Net::HTTP.start(uri.hostname, uri.port) do |http|
			http.request(req)
	end

	return res
end

def sendGetRequest(url) 
	uri = URI(url)
	req = Net::HTTP::Get.new(uri)
	res = Net::HTTP.start(uri.hostname, uri.port) do |http|
			http.request(req)
	end
	
	return res
end

def insertUser(body)
		url = 'http://0.0.0.0:3000/users'
		res = sendPostRequest(url, body)
	 
		return Integer(res.code)
end

def getUser(username)
		url = 'http://0.0.0.0:3000/users/' + username
		res = sendGetRequest(url)

		return Integer(res.code)
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

		body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
		@status_code = insertUser(body)
end

When("I try to retrieve user with {word}") do |username|
		body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
		if username == 'real'
			insertUser(body)
		end

		@status_code = getUser(username)
end

Then("I should see a {int} response") do |code|
		assert_equal(code, @status_code)
end

