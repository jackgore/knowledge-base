require 'net/http'
require 'test/unit'
require 'minitest/autorun'
require_relative 'util'

@status_code = 0

Given("I do have a running web server") do
end

When("I sign up with {word} credentials") do |isValid|
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			   username = 'UsersTest'
		else
			   username = 'a'
		end

		body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
		@status_code = Util.insertUser(body)
end

When("I try to retrieve user with {word}") do |username|
		body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
		if username == 'real'
			Util.insertUser(body)
		end

		@status_code = Util.getUser(username)
end

Then("I should see a {int} response") do |code|
		assert_equal(code, @status_code)
end

