require 'net/http'

module Util
	def Util.sendPostRequest(url, params)
		uri = URI(url)
		req = Net::HTTP::Post.new(uri, 'Content-Type' => 'application/json')
		req.body = params
		res = Net::HTTP.start(uri.hostname, uri.port) do |http|
				http.request(req)
		end

		return res
	end

	def Util.sendGetRequest(url) 
		uri = URI(url)
		req = Net::HTTP::Get.new(uri)
		res = Net::HTTP.start(uri.hostname, uri.port) do |http|
				http.request(req)
		end
		
		return res
	end
	
	def Util.login(body)
			url = 'http://0.0.0.0:3001/login'
			res = sendPostRequest(url, body)
		 
			return res
	end

	def Util.insertUser(body)
			url = 'http://0.0.0.0:3001/users'
			res = sendPostRequest(url, body)
		 
			return Integer(res.code)
	end
	
	def Util.insertTeam(body, orgName)
			url = 'http://0.0.0.0:3001/organizations/' + orgName + '/teams'
			res = sendPostRequest(url, body)
		 
			return Integer(res.code)
	end

	def Util.getUser(username)
			url = 'http://0.0.0.0:3001/users/' + username
			res = sendGetRequest(url)

			return Integer(res.code)
	end
	
	def Util.getOrganization(name)
			url = 'http://0.0.0.0:3001/organizations/' + name
			res = sendGetRequest(url)

			return Integer(res.code)
	end
	
	def Util.insertOrganization(body)
			url = 'http://0.0.0.0:3001/organizations'
			res = sendPostRequest(url, body)
		 
			return Integer(res.code)
	end
end
