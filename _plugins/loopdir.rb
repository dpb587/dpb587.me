# inspired by
# - https://gist.github.com/jgatjens/8925165
# - http://simon.heimlicher.com/articles/2012/02/01/jekyll-directory-listing

module Jekyll
  class Loopdir < Liquid::Block
    include Liquid::StandardFilters
    Syntax = /(#{Liquid::QuotedFragment}+)?/
 
    def initialize(tag_name, markup, tokens)
      @attributes = {}
 
      @attributes['path'] = nil;
      @attributes['parse'] = 'true';
      @attributes['match'] = '*';
      @attributes['sort'] = 'asc';
 
      markup.scan(Liquid::TagAttributes) do | key, value |
        @attributes[key] = value
      end

      if @attributes['path'].nil?
        raise SyntaxError.new("The `path` attribute is missing for `loopdir` tag.")
      end

      if 'true' == @attributes['parse']
        @attributes['parse'] = true
      else
        @attributes['parse'] = false
      end
 
      super
    end
 
    def render(context)
      context.registers[:loopdir] ||= Hash.new(0)
 
      files = Dir.glob(File.join(@attributes['path'], @attributes['match']))

      if @attributes['sort'].casecmp('desc') == 0
        files.sort! do | x, y |
          y <=> x
        end
      else
        files.sort!
      end
 
      result = []
 
      context.stack do
        files.each do |pathname|
          if @attributes['parse']
            data = {}

            content = File.read(pathname)

            if content =~ /^(---\s*\n.*?\n?)^(---\s*$\n?)/m
              content = $POSTMATCH

              begin
                data = YAML.load($1)
              rescue => e
                puts "YAML Exception reading #{name}: #{e.message}"
              end
            end

            data['name'] = File.basename(pathname, @attributes['match'].sub('*', ''))
            data['path'] = pathname
            data['content'] = content

            context['item'] = data
          else
            context['item'] = pathname
          end

          result << render_all(@nodelist, context)
        end
   
        result
      end
    end
  end
end
 
Liquid::Template.register_tag('loopdir', Jekyll::Loopdir)
