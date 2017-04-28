-- This script reads a single page of Markdown-like documentation and generates
-- output in three different forms: gofmt, git-flavored markdown, and
-- Caddy-style markdown.

local gsub, match, find, concat, insert = 
  string.gsub, string.match, string.find, table.concat, table.insert

local function write(filestr, str)
	local f = io.open(filestr, 'w+')
	if f then
		f:write(str)
		f:close()
	end
end

local function codeblock(tbl) 
	local newtbl = {}
	local incode = false
	for j, str in ipairs(tbl) do
		if find(str, '<code', 1, true) then
			incode = true
		elseif find(str, '</code', 1, true) then
			incode = false
		else
			if incode then
				str = '\t' .. str
			end
			insert(newtbl, str)
		end
	end
	return newtbl
end


local function caddywrite(tbl)
	local str = concat(tbl, '\n')
	-- str = gsub(str, '\n\n(%*%*.-)\n\n', '\n\n<mark class="block">%1</mark>\n\n')
	write('cgi.md', str)
end

local function readmewrite(tbl)
	tbl = codeblock(tbl)
	local str = concat(tbl, '\n')
	-- str = gsub(str, '%b<>', '')
	write('../README.md', str)
end

local function godocwrite(tbl)
	tbl = codeblock(tbl)
	for j, str in ipairs(tbl) do
		str = gsub(str, '^#+ *', '')
		tbl[j] = gsub(str, '^* ', '\n• ')
	end
	local str = concat(tbl, '\n')
	str = gsub(str, '\n\n\n+', '\n\n')
	str = gsub(str, '`', '')
	str = gsub(str, '/%*', '\x01')
	str = gsub(str, '%*', '')
	str = gsub(str, '\x01', '/*')
	-- str = gsub(str, '%b<>', '')
	-- replace [foo][bar] with foo
	str = gsub(str, '%[(%C-)%]%[%C-%]', '%1')
	str = '/*\n' .. str .. '\n*/\npackage cgi\n'
	write('../doc.go', str)
end

local godoc, caddy, readme = {}, {}, {}
local modeg, modec, moder

for str in io.lines('doc.txt') do
	local mode = string.match(str, '^~(%a+)~$')
	if mode then
		modeg = find(mode, 'g') ~= nil
		moder = find(mode, 'r') ~= nil
		modec = find(mode, 'c') ~= nil
	else
		if modeg then
			insert(godoc, str)
		end
		if modec then
			insert(caddy, str)
		end
		if moder then
			insert(readme, str)
		end
	end
end

caddywrite(caddy)
godocwrite(godoc)
readmewrite(readme)
