require 'net/http'
require 'test/unit'
require 'minitest/autorun'
require_relative 'util'

@status_code = 0
@org_valid = false
@org_name = 'TestOrganization'

When("I post a {word} organization") do |isValid|
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			name = 'TestOrganization'
			@org_name = 'TestOrganization'
		else
			@org_name = 'InvalidOrganization'
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

When("I post a {word} team") do |isValid|
		# Dirty hack - cucumber doesnt support booleans apparently
	  if isValid == 'true'
			name = 'TestTeam'
		else
			name = ''
		end

		body = {'name': name, organization: 1, 'is-public': true}.to_json
		@status_code = Util.insertTeam(body, @org_name)
end
