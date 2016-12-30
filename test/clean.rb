total = 0
Dir.glob('out/*') do |f|
  File.basename(f).length == "fbedda61ff8612d96d95542065666478e06307c785d82ae8fb8bd0c3f602cef7".length \
    or next
  File.unlink f
  total += 1
end
p total: total
