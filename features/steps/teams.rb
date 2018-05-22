require 'net/http'
require 'test/unit'
require 'minitest/autorun'
require_relative 'util'

@status_code = 0

When("I post a {word} organization") do |isValid|
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			   name = 'TestOrganization'
		else
			   name = ''
		end

		body = {'name': name, 'is-public': true}.to_json
		@status_code = Util.insertOrganization(body)
end

When("I try to retrieve organization with name {word}") do |name|
		body = {'name': name, 'is-public': true}.to_json
		if name != 'fake'
			Util.insertOrganization(body)
		end

		@status_code = Util.getOrganization(name)
end

