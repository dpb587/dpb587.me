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
      @attributes['sort'] = 'path';
 
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
 
      items = []
 
      Dir.glob(File.join(@attributes['path'], @attributes['match'])).each do |pathname|
        if @attributes['parse']
          item = {}

          content = File.read(pathname)

          if content =~ /^(---\s*\n.*?\n?)^(---\s*$\n?)/m
            content = $POSTMATCH

            begin
              item = YAML.load($1)
            rescue => e
              puts "YAML Exception reading #{name}: #{e.message}"
            end
          end

          item['content'] = content
        else
          context['item'] = pathname
        end

        item['name'] = File.basename(pathname, @attributes['match'].sub('*', ''))
        item['path'] = pathname

        items.push item
      end

      sortby = @attributes['sort'].gsub(/^-/, '')

      items.sort! do | x, y |
        x[sortby] <=> y[sortby]
      end

      if sortby != @attributes['sort']
        items.reverse!
      end

      context.stack do
        result = []

        items.each do | item |
          context['item'] = item
   
          result << render_all(@nodelist, context)
        end

        result
      end
    end
  end
end
 
Liquid::Template.register_tag('loopdir', Jekyll::Loopdir)
