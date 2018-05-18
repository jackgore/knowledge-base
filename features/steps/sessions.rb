require 'net/http'
require 'test/unit'
require 'minitest/autorun'
require_relative 'util'

Given("I do have an account") do
end

When("I am not already logged in") do
	# Just dont send a cookie when logging in
end

When("I login with {word} credentials") do |isValid|
		@valid = isValid
		username = 'Invalid'
		
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			username = 'SessionsTest'
			body = {'username': username, 'password': 'testpassword','first_name': 'Jack', 'last_name': 'Gore'}.to_json
			Util.insertUser(body) # Insert the user just to make sure its there TODO: enforce using same user/pass in each feature file
		end

		loginBody = {'username': username, 'password': 'testpassword'}.to_json
		@res = Util.login(loginBody)
end

Then("I should a see a Set-Cookie header") do
	if @valid == 'true'
		assert(@res.key?("Set-Cookie"))
		assert(@res["Set-Cookie"].is_a? String)
		refute_equal("", @res["Set-Cookie"])
	end
end

Then("I should see a {int} response to login") do |code|
		assert_equal(code, Integer(@res.code))
end

